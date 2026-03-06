package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"vibeta/internal/db"
)

// MessageService quản lý việc gửi và nhận messages qua Kafka
type MessageService struct {
	producer *Producer
	consumer *Consumer
	config   *ServiceConfig
}

// ServiceConfig cấu hình cho message service
type ServiceConfig struct {
	KafkaBrokers   []string
	MessageTopic   string
	ConsumerGroup  string
	WorkerCount    int
	EnableProducer bool
	EnableConsumer bool
}

// NewMessageService tạo một message service mới
func NewMessageService(database *db.Database) (*MessageService, error) {
	config := loadServiceConfig()

	service := &MessageService{
		config: config,
	}

	// Khởi tạo producer nếu được enable
	if config.EnableProducer {
		producerConfig := &ProducerConfig{
			Brokers: config.KafkaBrokers,
			Topic:   config.MessageTopic,
		}

		producer, err := NewProducer(producerConfig)
		if err != nil {
			return nil, fmt.Errorf("lỗi tạo Kafka producer: %w", err)
		}
		service.producer = producer
		log.Println("Kafka producer đã được khởi tạo")
	}

	// Khởi tạo consumer nếu được enable
	if config.EnableConsumer {
		consumerConfig := &ConsumerConfig{
			Brokers:       config.KafkaBrokers,
			Topic:         config.MessageTopic,
			ConsumerGroup: config.ConsumerGroup,
			WorkerCount:   config.WorkerCount,
		}

		consumer, err := NewConsumer(consumerConfig, database)
		if err != nil {
			return nil, fmt.Errorf("lỗi tạo Kafka consumer: %w", err)
		}
		service.consumer = consumer
		log.Println("Kafka consumer đã được khởi tạo")
	}

	return service, nil
}

// StartConsumer khởi động consumer để xử lý messages
func (ms *MessageService) StartConsumer(ctx context.Context) error {
	if ms.consumer == nil {
		return fmt.Errorf("consumer không được khởi tạo")
	}

	return ms.consumer.Start(ctx)
}

// GetProducer trả về Kafka producer
func (ms *MessageService) GetProducer() *Producer {
	return ms.producer
}

// GetConsumer trả về Kafka consumer
func (ms *MessageService) GetConsumer() *Consumer {
	return ms.consumer
}

// Close đóng tất cả connections
func (ms *MessageService) Close() error {
	var errors []string

	if ms.producer != nil {
		if err := ms.producer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("producer error: %v", err))
		}
	}

	if ms.consumer != nil {
		if err := ms.consumer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("consumer error: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing message service: %s", strings.Join(errors, ", "))
	}

	return nil
}

// loadServiceConfig load cấu hình từ environment variables
func loadServiceConfig() *ServiceConfig {
	config := &ServiceConfig{
		KafkaBrokers:   getEnvStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		MessageTopic:   getEnvString("KAFKA_MESSAGE_TOPIC", "chat_messages"),
		ConsumerGroup:  getEnvString("KAFKA_CONSUMER_GROUP", "chat_message_processors"),
		WorkerCount:    getEnvInt("KAFKA_WORKER_COUNT", 4),
		EnableProducer: getEnvBool("KAFKA_ENABLE_PRODUCER", true),
		EnableConsumer: getEnvBool("KAFKA_ENABLE_CONSUMER", true),
	}

	log.Printf("Kafka config loaded: brokers=%v, topic=%s, consumer_group=%s, workers=%d",
		config.KafkaBrokers, config.MessageTopic, config.ConsumerGroup, config.WorkerCount)

	return config
}

// Helper functions để đọc environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// HealthCheck kiểm tra trạng thái của Kafka connections
func (ms *MessageService) HealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"timestamp": time.Now(),
		"kafka": map[string]interface{}{
			"producer_enabled": ms.config.EnableProducer,
			"consumer_enabled": ms.config.EnableConsumer,
			"brokers":          ms.config.KafkaBrokers,
			"topic":            ms.config.MessageTopic,
		},
	}

	// TODO: Thêm logic kiểm tra connection thực tế

	return health
}
