package models

import "time"

// Response cấu trúc response chung
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginationRequest request phân trang
type PaginationRequest struct {
	Page     int `json:"page" form:"page" validate:"min=1"`
	PageSize int `json:"page_size" form:"page_size" validate:"min=1,max=100"`
}

// PaginationResponse response phân trang
type PaginationResponse struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Items      interface{} `json:"items"`
}

// ErrorCode mã lỗi
type ErrorCode string

const (
	ErrCodeValidation         ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeUserExists         ErrorCode = "USER_EXISTS"
	ErrCodeConversationExists ErrorCode = "CONVERSATION_EXISTS"
)

// APIError lỗi API
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

// OnlineUser người dùng đang online
type OnlineUser struct {
	UserID      string     `json:"user_id"`
	Username    string     `json:"username"`
	FullName    string     `json:"full_name"`
	Avatar      string     `json:"avatar,omitempty"`
	Status      UserStatus `json:"status"`
	LastActive  time.Time  `json:"last_active"`
	ConnectedAt time.Time  `json:"connected_at"`
}

// NotificationSettings cài đặt thông báo
type NotificationSettings struct {
	UserID                  string     `json:"user_id" bson:"user_id"`
	EnablePushNotification  bool       `json:"enable_push_notification" bson:"enable_push_notification"`
	EnableEmailNotification bool       `json:"enable_email_notification" bson:"enable_email_notification"`
	EnableSoundNotification bool       `json:"enable_sound_notification" bson:"enable_sound_notification"`
	MuteUntil               *time.Time `json:"mute_until,omitempty" bson:"mute_until,omitempty"`
}
