package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"vibeta/internal/db"
	"vibeta/internal/models"

	"github.com/IBM/sarama"
)

// Consumer xử lý messages từ Kafka queue
type Consumer struct {
	consumer       sarama.Consumer
	consumerGroup  sarama.ConsumerGroup
	config         *ConsumerConfig
	db             *db.Database
	processingPool *ProcessingPool
}

// ConsumerConfig cấu hình cho Kafka consumer
type ConsumerConfig struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	WorkerCount   int
}

// ProcessingPool quản lý workers để xử lý messages
type ProcessingPool struct {
	workers   int
	taskQueue chan *MessageEvent
	wg        sync.WaitGroup
	db        *db.Database
}

// MessageProcessor định nghĩa handler cho từng loại message
type MessageProcessor struct {
	db *db.Database
}

// NewConsumer tạo một Kafka consumer mới
func NewConsumer(config *ConsumerConfig, database *db.Database) (*Consumer, error) {
	// Cấu hình Sarama consumer
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	saramaConfig.Consumer.MaxProcessingTime = 1 * time.Minute
	saramaConfig.Consumer.Return.Errors = true

	// Tối ưu performance
	saramaConfig.Consumer.Fetch.Min = 1024 * 1024      // 1MB minimum fetch
	saramaConfig.Consumer.Fetch.Default = 1024 * 1024  // 1MB default fetch
	saramaConfig.Consumer.Fetch.Max = 10 * 1024 * 1024 // 10MB maximum fetch

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo consumer group: %w", err)
	}

	// Tạo processing pool
	pool := &ProcessingPool{
		workers:   config.WorkerCount,
		taskQueue: make(chan *MessageEvent, config.WorkerCount*10), // Buffer 10x số workers
		db:        database,
	}

	return &Consumer{
		consumerGroup:  consumerGroup,
		config:         config,
		db:             database,
		processingPool: pool,
	}, nil
}

// Start bắt đầu consumer để lắng nghe messages
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Bắt đầu Kafka consumer với %d workers", c.config.WorkerCount)

	// Khởi động processing pool
	c.processingPool.Start()

	// Khởi động error handler
	go c.handleErrors(ctx)

	// Khởi động consumer group
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := c.consumerGroup.Consume(ctx, []string{c.config.Topic}, c)
				if err != nil {
					log.Printf("Lỗi consumer group: %v", err)
				}
			}
		}
	}()

	log.Printf("Kafka consumer đã khởi động thành công")
	return nil
}

// Setup implements sarama.ConsumerGroupHandler
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group setup")
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group cleanup")
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Parse message event
			var event MessageEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Lỗi parse message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			// Đẩy vào processing pool
			select {
			case c.processingPool.taskQueue <- &event:
				// Message đã được đẩy vào queue để xử lý
			case <-session.Context().Done():
				return nil
			default:
				log.Printf("Processing pool đầy, dropping message %s", event.MessageID)
			}

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleErrors xử lý các lỗi từ consumer
func (c *Consumer) handleErrors(ctx context.Context) {
	for {
		select {
		case err := <-c.consumerGroup.Errors():
			if err != nil {
				log.Printf("Consumer error: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Start khởi động processing pool
func (p *ProcessingPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Stop dừng processing pool
func (p *ProcessingPool) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
}

// worker xử lý messages từ queue
func (p *ProcessingPool) worker(workerID int) {
	defer p.wg.Done()

	processor := &MessageProcessor{db: p.db}

	log.Printf("Worker %d đã khởi động", workerID)

	for event := range p.taskQueue {
		start := time.Now()

		if err := processor.ProcessEvent(event); err != nil {
			log.Printf("Worker %d: Lỗi xử lý message %s: %v", workerID, event.MessageID, err)
		} else {
			duration := time.Since(start)
			log.Printf("Worker %d: Đã xử lý message %s trong %v", workerID, event.MessageID, duration)
		}
	}

	log.Printf("Worker %d đã dừng", workerID)
}

// ProcessEvent xử lý một message event
func (mp *MessageProcessor) ProcessEvent(event *MessageEvent) error {
	switch event.Type {
	case "message":
		return mp.processMessage(event)
	case "reaction":
		return mp.processReaction(event)
	default:
		log.Printf("Không hỗ trợ message type: %s", event.Type)
		return nil
	}
}

// processMessage xử lý chat message
func (mp *MessageProcessor) processMessage(event *MessageEvent) error {
	message := &models.Message{
		ID:             event.MessageID,
		ConversationID: event.ConversationID,
		SenderID:       event.SenderID,
		Content:        event.Content,
		Type:           models.MessageType(event.MessageType),
		Status:         models.MessageStatusSent,
		CreatedAt:      event.Timestamp,
		UpdatedAt:      time.Now(),
	}

	err := mp.db.SaveMessage(message)
	if err != nil {
		return fmt.Errorf("lỗi lưu message vào DB: %w", err)
	}

	log.Printf("Đã lưu message %s vào database", event.MessageID)
	return nil
}

// processReaction xử lý reaction
func (mp *MessageProcessor) processReaction(event *MessageEvent) error {
	emoji, _ := event.Metadata["emoji"].(string)
	action, _ := event.Metadata["action"].(string)

	// TODO: Implement reaction processing
	// Hiện tại chỉ log để tracking
	log.Printf("Xử lý reaction: user %s %s emoji %s cho message %s",
		event.SenderID, action, emoji, event.MessageID)

	return nil
}

// Close đóng consumer
func (c *Consumer) Close() error {
	log.Println("Đang đóng Kafka consumer...")

	// Stop processing pool
	c.processingPool.Stop()

	// Close consumer group
	if err := c.consumerGroup.Close(); err != nil {
		return fmt.Errorf("lỗi đóng consumer group: %w", err)
	}

	log.Println("Kafka consumer đã đóng")
	return nil
}
