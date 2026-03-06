package kafka

import (
	"encoding/json"
	"log"
	"time"

	"vibeta/internal/models"

	"github.com/IBM/sarama"
)

// Producer cung cấp interface để gửi messages vào Kafka
type Producer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
}

// ProducerConfig cấu hình cho Kafka producer
type ProducerConfig struct {
	Brokers []string
	Topic   string
}

// MessageEvent định nghĩa cấu trúc message sẽ được gửi qua Kafka
type MessageEvent struct {
	Type           string                 `json:"type"`
	MessageID      string                 `json:"message_id"`
	ConversationID string                 `json:"conversation_id"`
	SenderID       string                 `json:"sender_id"`
	Content        string                 `json:"content"`
	MessageType    string                 `json:"message_type"`
	Reactions      map[string][]string    `json:"reactions,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// NewProducer tạo một Kafka producer mới
func NewProducer(config *ProducerConfig) (*Producer, error) {
	// Cấu hình Sarama
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll // Đợi confirmation từ tất cả replicas
	saramaConfig.Producer.Retry.Max = 3                    // Retry tối đa 3 lần
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	// Cải thiện performance và reliability
	saramaConfig.Producer.Flush.Frequency = 100 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 100
	saramaConfig.Producer.MaxMessageBytes = 1000000

	// Compression để giảm network traffic
	saramaConfig.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// PublishMessage gửi một message event vào Kafka queue
func (p *Producer) PublishMessage(event *MessageEvent) error {
	// Serialize message event thành JSON
	messageBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Tạo Kafka message
	msg := &sarama.ProducerMessage{
		Topic: p.config.Topic,
		Key:   sarama.StringEncoder(event.MessageID), // Sử dụng message_id làm key để partition
		Value: sarama.ByteEncoder(messageBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(event.Type),
			},
			{
				Key:   []byte("conversation_id"),
				Value: []byte(event.ConversationID),
			},
		},
		Timestamp: event.Timestamp,
	}

	// Gửi message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Lỗi gửi message vào Kafka: %v", err)
		return err
	}

	log.Printf("Message %s được gửi thành công vào partition %d với offset %d",
		event.MessageID, partition, offset)
	return nil
}

// PublishReaction gửi một reaction event vào Kafka queue
func (p *Producer) PublishReaction(messageID, userID, emoji, action string, conversationID string) error {
	event := &MessageEvent{
		Type:           "reaction",
		MessageID:      messageID,
		ConversationID: conversationID,
		SenderID:       userID,
		MessageType:    "reaction",
		Metadata: map[string]interface{}{
			"emoji":  emoji,
			"action": action, // "add" hoặc "remove"
		},
		Timestamp: time.Now(),
	}

	return p.PublishMessage(event)
}

// PublishChatMessage gửi một chat message vào Kafka queue
func (p *Producer) PublishChatMessage(wsMsg models.WebSocketMessage, userID string) error {
	data, ok := wsMsg.Data.(map[string]interface{})
	if !ok {
		log.Printf("Lỗi parse message data")
		return nil // Không return error để không block WebSocket
	}

	content, _ := data["content"].(string)
	messageType, _ := data["type"].(string)
	messageID, _ := data["message_id"].(string)

	// Tạo ID nếu chưa có
	if messageID == "" {
		messageID = generateMessageID(userID)
	}

	event := &MessageEvent{
		Type:           "message",
		MessageID:      messageID,
		ConversationID: wsMsg.ConvID,
		SenderID:       userID,
		Content:        content,
		MessageType:    messageType,
		Timestamp:      time.Now(),
	}

	return p.PublishMessage(event)
}

// Close đóng Kafka producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// Helper function để tạo message ID
func generateMessageID(userID string) string {
	return "msg_" + userID + "_" + time.Now().Format("20060102150405") + "_" + time.Now().Format("000")
}
