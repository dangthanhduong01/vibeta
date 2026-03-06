package db

import (
	"log"
	"vibeta/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

// NewDatabase tạo kết nối database mới
func NewDatabase() *Database {
	// Kết nối đến PostgreSQL
	// TODO: Thay đổi connection string theo cấu hình của bạn
	// Hiện tại sử dụng SQLite cho demo đơn giản
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=postgres dbname=vibeta_chat port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		// Fallback to SQLite for development
		log.Println("Không thể kết nối PostgreSQL, sử dụng SQLite cho development")
		return NewSQLiteDatabase()
	}

	// Auto migrate các tables
	err = db.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
		&models.ConversationParticipant{},
	)

	if err != nil {
		log.Fatal("Không thể migrate database:", err)
	}

	log.Println("PostgreSQL database đã kết nối và migrate thành công")

	return &Database{DB: db}
}

// NewSQLiteDatabase tạo database SQLite cho development
func NewSQLiteDatabase() *Database {
	db, err := gorm.Open(sqlite.Open("vibeta_chat.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Không thể tạo SQLite database:", err)
	}

	// Auto migrate các tables
	err = db.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
		&models.ConversationParticipant{},
	)

	if err != nil {
		log.Fatal("Không thể migrate SQLite database:", err)
	}

	log.Println("SQLite database đã được tạo và migrate thành công")

	return &Database{DB: db}
}

// SaveMessage lưu tin nhắn vào database
func (d *Database) SaveMessage(message *models.Message) error {
	return d.DB.Create(message).Error
}

// GetMessages lấy tin nhắn từ database theo conversation ID
func (d *Database) GetMessages(conversationID string, limit int, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := d.DB.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, err
}

// SaveConversation lưu cuộc trò chuyện vào database
func (d *Database) SaveConversation(conv *models.Conversation) error {
	return d.DB.Create(conv).Error
}

// GetConversations lấy danh sách cuộc trò chuyện của user
func (d *Database) GetConversations(userID string) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := d.DB.Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Find(&conversations).Error

	return conversations, err
}

// SaveUser lưu thông tin user
func (d *Database) SaveUser(user *models.User) error {
	return d.DB.Create(user).Error
}

// GetUser lấy thông tin user
func (d *Database) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := d.DB.Where("id = ?", userID).First(&user).Error
	return &user, err
}

// AddParticipantToConversation thêm participant vào conversation
func (d *Database) AddParticipantToConversation(conversationID, userID string) error {
	participant := &models.ConversationParticipant{
		ConversationID: conversationID,
		UserID:         userID,
	}
	return d.DB.Create(participant).Error
}

// UpdateMessage cập nhật tin nhắn (cho reactions)
func (d *Database) UpdateMessage(messageID string, message *models.Message) error {
	return d.DB.Where("id = ?", messageID).Updates(message).Error
}
