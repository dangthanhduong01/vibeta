package models

import (
	"time"
)

// User đại diện cho một người dùng trong hệ thống
type User struct {
	ID         string     `json:"id" bson:"_id,omitempty"`
	Username   string     `json:"username" bson:"username"`
	Email      string     `json:"email" bson:"email"`
	FullName   string     `json:"full_name" bson:"full_name"`
	Avatar     string     `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Status     UserStatus `json:"status" bson:"status"`
	LastActive time.Time  `json:"last_active" bson:"last_active"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" bson:"updated_at"`
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
