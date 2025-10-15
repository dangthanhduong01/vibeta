package models

import (
	"time"
)

// MessageType đại diện cho loại tin nhắn
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeSystem   MessageType = "system" // Tin nhắn hệ thống (user join, leave, etc.)
	MessageTypeReaction MessageType = "reaction"
)

// MessageStatus trạng thái tin nhắn
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// Message đại diện cho một tin nhắn
type Message struct {
	ID             string        `json:"id" bson:"_id,omitempty"`
	ConversationID string        `json:"conversation_id" bson:"conversation_id"`
	SenderID       string        `json:"sender_id" bson:"sender_id"`
	Content        string        `json:"content" bson:"content"`
	Type           MessageType   `json:"type" bson:"type"`
	Status         MessageStatus `json:"status" bson:"status"`
	ReplyToID      string        `json:"reply_to_id,omitempty" bson:"reply_to_id,omitempty"` // ID tin nhắn được reply
	Attachments    []Attachment  `json:"attachments,omitempty" bson:"attachments,omitempty"`
	Reactions      []Reaction    `json:"reactions,omitempty" bson:"reactions,omitempty"`
	EditedAt       *time.Time    `json:"edited_at,omitempty" bson:"edited_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" bson:"updated_at"`
}

// Attachment đính kèm file
type Attachment struct {
	ID       string `json:"id" bson:"id"`
	FileName string `json:"file_name" bson:"file_name"`
	FileSize int64  `json:"file_size" bson:"file_size"`
	FileType string `json:"file_type" bson:"file_type"`
	URL      string `json:"url" bson:"url"`
}

// Reaction phản ứng với tin nhắn
type Reaction struct {
	UserID string `json:"user_id" bson:"user_id"`
	Emoji  string `json:"emoji" bson:"emoji"`
}

// SendMessageRequest request gửi tin nhắn
type SendMessageRequest struct {
	ConversationID string       `json:"conversation_id" validate:"required"`
	Content        string       `json:"content" validate:"required"`
	Type           MessageType  `json:"type" validate:"required"`
	ReplyToID      string       `json:"reply_to_id,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
}

// MessageWithSender tin nhắn kèm thông tin người gửi
type MessageWithSender struct {
	Message
	Sender User `json:"sender"`
}

// WebSocketMessage tin nhắn qua WebSocket
type WebSocketMessage struct {
	Type   string      `json:"type"` // "message", "typing", "user_joined", "user_left", etc.
	Data   interface{} `json:"data"`
	UserID string      `json:"user_id,omitempty"`
	ConvID string      `json:"conversation_id,omitempty"`
}

// TypingIndicator chỉ báo đang gõ
type TypingIndicator struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	IsTyping       bool   `json:"is_typing"`
}
