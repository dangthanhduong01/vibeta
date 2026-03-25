# VibeTA - Scalable Real-time Chat Application

### Features

- **Real-time messaging**: WebSocket connections with low latency
- **Emoji reactions**: React message with emojis  
- **Scalable architecture**: Kafka-based message queue
- **Database persistence**: PostgreSQL with GORM ORM
- **Graceful shutdown**: Production-ready with health checks
- **Monitoring**: Kafka UI & application metrics

### Quick Start

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

