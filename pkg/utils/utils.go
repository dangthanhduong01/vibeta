package utils

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string `mapstructure:"HOST"`
	User     string ``
	Password string ``
	DbName   string ``
	Port     string ``
}

func InitDB() {
	dsn := "host=localhost user=postgres password=yourpassword dbname=library port=5432 sslmode=disable"
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	fmt.Println("✅ Kết nối PostgreSQL thành công!")

	db.AutoMigrate()
}
