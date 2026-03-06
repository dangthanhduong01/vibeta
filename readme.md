# VibeTA - Scalable Real-time Chat Application

## 🚀 Kiến trúc mới với Kafka Message Queue

VibeTA đã được nâng cấp thành hệ thống chat scalable với kiến trúc microservices và Kafka message queue.

### ✨ Tính năng chính

- **Real-time messaging**: WebSocket connections với low latency
- **Emoji reactions**: React tin nhắn với emojis  
- **Scalable architecture**: Kafka-based message queue
- **Database persistence**: PostgreSQL với GORM ORM
- **Graceful shutdown**: Production-ready với health checks
- **Monitoring**: Kafka UI và application metrics

### 🚀 Quick Start

```bash
# 1. Start infrastructure
make docker-up && make kafka-topics

# 2. Build and run
make build && make run-all

# 3. Access application  
# Chat: http://localhost:8080
# Kafka UI: http://localhost:8090
```

### 📋 Available Commands

```bash
make help           # Show all commands
make docker-up      # Start Kafka, PostgreSQL  
make build          # Build binaries
make run-all        # Run WebSocket + Workers
make health         # Check system health
```

📖 **Full documentation**: [SCALABLE_ARCHITECTURE.md](SCALABLE_ARCHITECTURE.md)

---

## Appchat websocket golang

