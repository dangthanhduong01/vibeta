package models

import (
	"time"

	"gorm.io/gorm"
)

// User đại diện cho một người dùng trong hệ thống
type User struct {
	ID         string         `json:"id" gorm:"primaryKey"`
	Username   string         `json:"username" gorm:"uniqueIndex;not null"`
	Email      string         `json:"email" gorm:"uniqueIndex;not null"`
	FullName   string         `json:"full_name" gorm:"not null"`
	Avatar     string         `json:"avatar,omitempty"`
	Status     UserStatus     `json:"status" gorm:"default:offline"`
	LastActive time.Time      `json:"last_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserStatus đại diện cho trạng thái của người dùng
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusOffline UserStatus = "offline"
	UserStatusAway    UserStatus = "away"
	UserStatusBusy    UserStatus = "busy"
)

// CreateUserRequest request tạo user mới
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=50"`
	Avatar   string `json:"avatar,omitempty"`
}

// UpdateUserRequest request cập nhật user
type UpdateUserRequest struct {
	FullName string     `json:"full_name,omitempty"`
	Avatar   string     `json:"avatar,omitempty"`
	Status   UserStatus `json:"status,omitempty"`
}
