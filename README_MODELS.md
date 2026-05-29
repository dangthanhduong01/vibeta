# ViBeta - Hệ thống Chat WebSocket với Go

## Tổng quan
ViBeta là một hệ thống chat real-time sử dụng WebSocket được xây dựng bằng Go. Hệ thống hỗ trợ chat 1-1 và chat nhóm với các tính năng hiện đại.

## Cấu trúc Models

### 1. User Model (`internal/models/users.go`)
```go
type User struct {
    ID          string     `json:"id"`
    Username    string     `json:"username"`
    Email       string     `json:"email"`
    FullName    string     `json:"full_name"`
    Avatar      string     `json:"avatar,omitempty"`
    Status      UserStatus `json:"status"`
    LastActive  time.Time  `json:"last_active"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

**Trạng thái User:**
- `online`: Đang online
- `offline`: Offline
- `away`: Xa máy
- `busy`: Bận

### 2. Conversation Model (`internal/models/conversation.go`)
```go
type Conversation struct {
    ID           string           `json:"id"`
    Type         ConversationType `json:"type"`
    Name         string           `json:"name,omitempty"`
    Description  string           `json:"description,omitempty"`
    Avatar       string           `json:"avatar,omitempty"`
    Participants []string         `json:"participants"`
    CreatedBy    string           `json:"created_by"`
    LastMessage  *LastMessage     `json:"last_message,omitempty"`
    CreatedAt    time.Time        `json:"created_at"`
    UpdatedAt    time.Time        `json:"updated_at"`
}
```

**Loại Conversation:**
- `direct`: Chat 1-1 giữa 2 người
- `group`: Chat nhóm (nhiều người)

### 3. Message Model (`internal/models/message.go`)
```go
type Message struct {
    ID             string        `json:"id"`
    ConversationID string        `json:"conversation_id"`
    SenderID       string        `json:"sender_id"`
    Content        string        `json:"content"`
    Type           MessageType   `json:"type"`
    Status         MessageStatus `json:"status"`
    ReplyToID      string        `json:"reply_to_id,omitempty"`
    Attachments    []Attachment  `json:"attachments,omitempty"`
    Reactions      []Reaction    `json:"reactions,omitempty"`
    EditedAt       *time.Time    `json:"edited_at,omitempty"`
    CreatedAt      time.Time     `json:"created_at"`
    UpdatedAt      time.Time     `json:"updated_at"`
}
```

**Loại Message:**
- `text`: Tin nhắn văn bản
- `image`: Hình ảnh
- `file`: File đính kèm
- `system`: Tin nhắn hệ thống
- `reaction`: Phản ứng

**Trạng thái Message:**
- `sent`: Đã gửi
- `delivered`: Đã nhận
- `read`: Đã đọc
- `failed`: Gửi thất bại

## Cấu trúc WebSocket

### WebSocket Message Format
```go
type WebSocketMessage struct {
    Type    string      `json:"type"`
    Data    interface{} `json:"data"`
    UserID  string      `json:"user_id,omitempty"`
    ConvID  string      `json:"conversation_id,omitempty"`
}
```

### Các loại WebSocket message:
1. **join_conversation**: Tham gia cuộc trò chuyện
2. **leave_conversation**: Rời cuộc trò chuyện
3. **message**: Gửi tin nhắn
4. **typing**: Thông báo đang gõ

## Cách chạy ứng dụng

### 1. Cài đặt dependencies
```bash
go mod tidy
```

### 2. Chạy server
```bash
cd ws
go run main.go
```

### 3. Mở trình duyệt
Truy cập: `http://localhost:8080`

## Tính năng

### Đã triển khai:
- WebSocket connection management
- Real-time messaging
- Conversation management
- User management trong WebSocket
- Typing indicators
- Join/Leave conversations

### Sẽ triển khai:
- Database integration (MongoDB/PostgreSQL)
- User authentication & authorization
- File upload & sharing
- Message reactions
- Message editing & deletion
- Push notifications
- Message search
- Message history pagination

## API Endpoints (Tương lai)

### User APIs:
- `POST /api/users` - Tạo user mới
- `GET /api/users/{id}` - Lấy thông tin user
- `PUT /api/users/{id}` - Cập nhật user
- `GET /api/users/me` - Lấy thông tin user hiện tại

### Conversation APIs:
- `POST /api/conversations` - Tạo conversation mới
- `GET /api/conversations` - Lấy danh sách conversations
- `GET /api/conversations/{id}` - Lấy thông tin conversation
- `PUT /api/conversations/{id}` - Cập nhật conversation
- `POST /api/conversations/{id}/participants` - Thêm người tham gia
- `DELETE /api/conversations/{id}/participants/{userId}` - Xóa người tham gia

### Message APIs:
- `GET /api/conversations/{id}/messages` - Lấy tin nhắn trong conversation
- `POST /api/conversations/{id}/messages` - Gửi tin nhắn mới
- `PUT /api/messages/{id}` - Chỉnh sửa tin nhắn
- `DELETE /api/messages/{id}` - Xóa tin nhắn

## Cấu trúc thư mục

```
vibeta/
├── ws/
│   └── main.go                 # WebSocket server
├── internal/
│   └── models/
│       ├── users.go           # User models
│       ├── conversation.go    # Conversation models
│       ├── message.go         # Message models
│       └── common.go          # Common types
├── client/
│   └── index.html             # Web client
├── cmd/                       # Future API server
├── go.mod
└── README.md
```

## Demo

Server sẽ chạy trên `localhost:8080` với:
- WebSocket endpoint: `/ws`
- Static file serving: `/` (phục vụ `index.html`)

Mở nhiều tab trình duyệt để test chat real-time!
