package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeSystem   MessageType = "system" // Tin nhắn hệ thống (user join, leave, etc.)
	MessageTypeReaction MessageType = "reaction"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type Message struct {
	ID             string         `json:"id" gorm:"primaryKey"`
	ConversationID string         `json:"conversation_id" gorm:"not null;index"`
	SenderID       string         `json:"sender_id" gorm:"not null;index"`
	Content        string         `json:"content"`
	Type           MessageType    `json:"type" gorm:"not null"`
	Status         MessageStatus  `json:"status" gorm:"default:sent"`
	ReplyToID      string         `json:"reply_to_id,omitempty" gorm:"index"`
	Attachments    string         `json:"attachments,omitempty" gorm:"type:text"` // JSON string
	Reactions      string         `json:"reactions,omitempty" gorm:"type:text"`   // JSON string
	EditedAt       *time.Time     `json:"edited_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Sender       User         `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Conversation Conversation `json:"conversation,omitempty" gorm:"foreignKey:ConversationID"`
}

type Attachment struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
	URL      string `json:"url"`
}

type Reaction struct {
	UserID string `json:"user_id"`
	Emoji  string `json:"emoji"`
}

type SendMessageRequest struct {
	ConversationID string       `json:"conversation_id" validate:"required"`
	Content        string       `json:"content" validate:"required"`
	Type           MessageType  `json:"type" validate:"required"`
	ReplyToID      string       `json:"reply_to_id,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
}

type MessageWithSender struct {
	Message
	Sender User `json:"sender"`
}

type WebSocketMessage struct {
	Type   string      `json:"type"` // "message", "typing", "user_joined", "user_left", etc.
	Data   interface{} `json:"data"`
	UserID string      `json:"user_id,omitempty"`
	ConvID string      `json:"conversation_id,omitempty"`
}

type TypingIndicator struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	IsTyping       bool   `json:"is_typing"`
}
