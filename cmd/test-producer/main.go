package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"vibeta/internal/kafka"
	"vibeta/internal/models"
)

func main() {
	log.Println("Testing Kafka Producer...")

	// Test config
	config := &kafka.ProducerConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "chat_messages",
	}

	// Create producer
	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Test message
	wsMsg := models.WebSocketMessage{
		Type:   "message",
		ConvID: "test_conversation",
		Data: map[string]interface{}{
			"content":    "Hello from Kafka test!",
			"type":       "text",
			"message_id": "test_msg_001",
		},
	}

	// Send message
	err = producer.PublishChatMessage(wsMsg, "test_user")
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	} else {
		log.Println("Message sent successfully!")
	}

	// Test reaction
	err = producer.PublishReaction("test_msg_001", "test_user", "😀", "add", "test_conversation")
	if err != nil {
		log.Printf("Failed to send reaction: %v", err)
	} else {
		log.Println("Reaction sent successfully!")
	}

	// Test with custom event
	event := &kafka.MessageEvent{
		Type:           "custom_event",
		MessageID:      "custom_001",
		ConversationID: "test_conversation",
		SenderID:       "test_user",
		Content:        "Custom event content",
		MessageType:    "custom",
		Metadata: map[string]interface{}{
			"custom_field": "custom_value",
		},
		Timestamp: time.Now(),
	}

	err = producer.PublishMessage(event)
	if err != nil {
		log.Printf("Failed to send custom event: %v", err)
	} else {
		log.Println("Custom event sent successfully!")
	}

	// Health check simulation
	healthData := map[string]interface{}{
		"status":    "testing",
		"timestamp": time.Now(),
		"producer":  "healthy",
	}

	healthBytes, _ := json.Marshal(healthData)
	fmt.Printf("Health check data: %s\n", string(healthBytes))

	log.Println("Kafka producer test completed!")
}
