# VibeTA Chat - Scalable Architecture với Kafka

## Tổng quan kiến trúc mới

Hệ thống chat đã được nâng cấp thành kiến trúc scalable với Kafka message queue:

```
[WebSocket Client] 
       ↓
[WebSocket Server] → [Kafka Producer] → [Kafka Queue] 
                                             ↓
[Real-time Broadcast]              [Message Worker] 
                                             ↓
                                      [Database]
```

### Các thành phần chính:

1. **WebSocket Server** (`ws/main.go`)
   - Xử lý kết nối real-time từ clients
   - Gửi messages vào Kafka queue thay vì lưu trực tiếp DB
   - Broadcast tin nhắn real-time đến clients
   - Graceful shutdown và health checks

2. **Message Worker** (`cmd/worker/main.go`) 
   - Consumer Kafka messages từ queue
   - Xử lý và lưu messages vào database
   - Hỗ trợ multiple workers để xử lý song song
   - Auto-scaling với worker pool

3. **Kafka Message Queue**
   - Topic: `chat_messages`
   - Partitions: 3 (có thể scale thêm)
   - Message persistence và replication

## Cài đặt và chạy hệ thống

### 1. Khởi động infrastructure (Kafka, PostgreSQL)

```bash
# Khởi động Docker containers
make docker-up

# Tạo Kafka topics
make kafka-topics
```

### 2. Build và chạy ứng dụng

```bash
# Build binaries
make build

# Chạy cả WebSocket server và Workers
make run-all
```

### 3. Hoặc chạy từng service riêng

```bash
# Terminal 1 - Chạy Message Worker
make run-worker

# Terminal 2 - Chạy WebSocket Server  
make run-websocket
```

## Cấu hình

### Environment Variables (.env)

```bash
# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_MESSAGE_TOPIC=chat_messages
KAFKA_CONSUMER_GROUP=chat_message_processors
KAFKA_WORKER_COUNT=4

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=vibeta_chat
```

### Scaling Workers

Điều chỉnh số lượng workers:
```bash
export KAFKA_WORKER_COUNT=8  # Tăng thành 8 workers
make run-worker
```

## Monitoring và Debug

### 1. Health Check API
```bash
curl http://localhost:8080/health
```

### 2. Kafka UI
Truy cập: http://localhost:8090
- Xem topics, partitions
- Monitor message throughput
- Consumer lag tracking

### 3. Application Logs
```bash
# Xem logs của Worker
make logs-worker

# Xem logs của Kafka
make logs-kafka

# Xem logs của PostgreSQL
make logs-postgres
```

## Ưu điểm của kiến trúc mới

### 1. **Scalability**
- WebSocket servers có thể scale độc lập
- Workers có thể scale theo throughput
- Kafka handle millions messages/second

### 2. **Reliability** 
- Messages không bị mất khi server crash
- Kafka persistence và replication
- Graceful shutdown và error handling

### 3. **Performance**
- WebSocket response nhanh (không đợi DB write)
- Batch processing với workers
- Asynchronous message processing

### 4. **Separation of Concerns**
- Real-time communication (WebSocket)
- Message persistence (Workers) 
- Business logic tách biệt

## Message Flow

### 1. Chat Message
```
Client → WebSocket → Kafka Producer → Queue → Worker → Database
                  ↓
               Real-time Broadcast
```

### 2. Emoji Reactions  
```
Client → WebSocket → Kafka Producer → Queue → Worker → Update DB
                  ↓
               Real-time Broadcast
```

## Load Testing

Test với multiple clients:
```bash
# Terminal 1: Start system
make start

# Terminal 2: Send test messages
for i in {1..1000}; do
  curl -X POST http://localhost:8080/test-message -d "message=$i"
  sleep 0.1
done
```

Monitor trong Kafka UI để xem throughput.

## Troubleshooting

### 1. Kafka Connection Issues
```bash
# Check Kafka status
docker exec vibeta-kafka kafka-broker-api-versions --bootstrap-server localhost:9092

# Check topics
docker exec vibeta-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### 2. Database Connection Issues  
```bash
# Check PostgreSQL
docker exec vibeta-postgres pg_isready -U postgres

# Connect to database
docker exec -it vibeta-postgres psql -U postgres -d vibeta_chat
```

### 3. High Consumer Lag
- Tăng số workers: `KAFKA_WORKER_COUNT=8`
- Tăng số partitions của topic
- Optimize database queries

## Performance Tuning

### 1. Kafka Producer
- Batch size: Điều chỉnh trong `producer.go`
- Compression: Sử dụng Snappy
- Async writes với callbacks

### 2. Kafka Consumer
- Fetch size optimization
- Multiple workers per consumer group
- Commit strategy tuning

### 3. Database
- Connection pooling
- Batch inserts
- Proper indexing

## Deployment

### Development
```bash
make dev  # Live reload với Air
```

### Production  
```bash
make deploy  # Setup toàn bộ infrastructure
```

### Docker Production
```bash
# Build production images
docker build -t vibeta-websocket -f Dockerfile.websocket .
docker build -t vibeta-worker -f Dockerfile.worker .

# Deploy with docker-compose.prod.yml
docker-compose -f docker-compose.prod.yml up -d
```

## Next Steps

1. **Redis Integration**: Session management và caching
2. **Metrics**: Prometheus + Grafana monitoring  
3. **Load Balancing**: Multiple WebSocket server instances
4. **Database Sharding**: Scale PostgreSQL horizontally
5. **Message Encryption**: End-to-end encryption cho messages
6. **Rate Limiting**: Prevent spam và abuse
