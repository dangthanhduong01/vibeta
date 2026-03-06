package models

import (
	"time"

	"gorm.io/gorm"
)

// ConversationType đại diện cho loại cuộc trò chuyện
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct" // Chat 1-1
	ConversationTypeGroup  ConversationType = "group"  // Chat nhóm
)

// Conversation đại diện cho một cuộc trò chuyện
type Conversation struct {
	ID          string           `json:"id" gorm:"primaryKey"`
	Type        ConversationType `json:"type" gorm:"not null"`
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Avatar      string           `json:"avatar,omitempty"`
	CreatedBy   string           `json:"created_by" gorm:"not null;index"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relations
	Creator      User                      `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Participants []ConversationParticipant `json:"participants,omitempty" gorm:"foreignKey:ConversationID"`
	Messages     []Message                 `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

// ConversationParticipant người tham gia cuộc trò chuyện
type ConversationParticipant struct {
	ID             uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	ConversationID string     `json:"conversation_id" gorm:"not null;index"`
	UserID         string     `json:"user_id" gorm:"not null;index"`
	JoinedAt       time.Time  `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	LeftAt         *time.Time `json:"left_at,omitempty"`

	// Relations
	Conversation Conversation `json:"conversation,omitempty" gorm:"foreignKey:ConversationID"`
	User         User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// LastMessage thông tin tin nhắn cuối cùng trong conversation
type LastMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	SenderID  string    `json:"sender_id"`
	Timestamp time.Time `json:"timestamp"`
}

// CreateConversationRequest request tạo cuộc trò chuyện mới
type CreateConversationRequest struct {
	Type           ConversationType `json:"type" validate:"required"`
	Name           string           `json:"name,omitempty" validate:"max=100"`
	Description    string           `json:"description,omitempty" validate:"max=500"`
	ParticipantIDs []string         `json:"participant_ids" validate:"required,min=1"`
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
