package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vibeta/internal/db"
	"vibeta/internal/kafka"
)

func main() {
	log.Println("Starting Message Worker Service...")

	// Khởi tạo database
	database := db.NewDatabase()
	if database == nil {
		log.Fatal("Không thể kết nối database")
	}

	// Khởi tạo Kafka message service (chỉ consumer)
	os.Setenv("KAFKA_ENABLE_PRODUCER", "false") // Worker chỉ cần consumer
	os.Setenv("KAFKA_ENABLE_CONSUMER", "true")

	messageService, err := kafka.NewMessageService(database)
	if err != nil {
		log.Fatalf("Lỗi khởi tạo message service: %v", err)
	}
	defer messageService.Close()

	// Tạo context với graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Khởi động consumer
	if err := messageService.StartConsumer(ctx); err != nil {
		log.Fatalf("Lỗi khởi động consumer: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Message Worker Service đã khởi động. Đang lắng nghe messages...")
	log.Println("Nhấn Ctrl+C để dừng...")

	// Chờ signal để shutdown
	<-sigChan
	log.Println("Nhận được signal shutdown...")

	// Graceful shutdown với timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Cancel main context
	cancel()

	// Đợi service shutdown hoàn tất
	done := make(chan bool, 1)
	go func() {
		messageService.Close()
		done <- true
	}()

	select {
	case <-done:
		log.Println("Message Worker Service đã shutdown thành công")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout, force exit")
	}
}
