package models

import (
	"time"
)

// ConversationType đại diện cho loại cuộc trò chuyện
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct" // Chat 1-1
	ConversationTypeGroup  ConversationType = "group"  // Chat nhóm
)

// Conversation đại diện cho một cuộc trò chuyện
type Conversation struct {
	ID           string           `json:"id" bson:"_id,omitempty"`
	Type         ConversationType `json:"type" bson:"type"`
	Name         string           `json:"name,omitempty" bson:"name,omitempty"`               // Tên nhóm (chỉ dùng cho group)
	Description  string           `json:"description,omitempty" bson:"description,omitempty"` // Mô tả nhóm
	Avatar       string           `json:"avatar,omitempty" bson:"avatar,omitempty"`           // Avatar nhóm
	Participants []string         `json:"participants" bson:"participants"`                   // Danh sách ID của người tham gia
	CreatedBy    string           `json:"created_by" bson:"created_by"`                       // ID người tạo
	LastMessage  *LastMessage     `json:"last_message,omitempty" bson:"last_message,omitempty"`
	CreatedAt    time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" bson:"updated_at"`
}

// LastMessage thông tin tin nhắn cuối cùng trong conversation
type LastMessage struct {
	ID        string    `json:"id" bson:"id"`
	Content   string    `json:"content" bson:"content"`
	Type      string    `json:"type" bson:"type"`
	SenderID  string    `json:"sender_id" bson:"sender_id"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// ConversationParticipant thông tin người tham gia cuộc trò chuyện
type ConversationParticipant struct {
	ConversationID string    `json:"conversation_id" bson:"conversation_id"`
	UserID         string    `json:"user_id" bson:"user_id"`
	Role           string    `json:"role" bson:"role"` // "admin", "member"
	JoinedAt       time.Time `json:"joined_at" bson:"joined_at"`
	LastReadAt     time.Time `json:"last_read_at" bson:"last_read_at"`
}

// CreateConversationRequest request tạo cuộc trò chuyện mới
type CreateConversationRequest struct {
	Type         ConversationType `json:"type" validate:"required"`
	Name         string           `json:"name,omitempty" validate:"max=100"`
	Description  string           `json:"description,omitempty" validate:"max=500"`
	Participants []string         `json:"participants" validate:"required,min=1"`
}

// AddParticipantRequest request thêm người tham gia
type AddParticipantRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1"`
}

// UpdateConversationRequest request cập nhật cuộc trò chuyện
type UpdateConversationRequest struct {
	Name        string `json:"name,omitempty" validate:"max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	Avatar      string `json:"avatar,omitempty"`
}

// ConversationWithDetails conversation với thông tin chi tiết
type ConversationWithDetails struct {
	Conversation
	ParticipantDetails []User `json:"participant_details"`
	UnreadCount        int    `json:"unread_count"`
}
