package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"vibeta/internal/models"

	"github.com/gorilla/websocket"
)

// Upgrader chuyển đổi một kết nối HTTP thông thường thành một kết nối WebSocket.
// Chúng ta cần chỉ định kích thước bộ đệm đọc và ghi.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Chúng ta cần kiểm tra nguồn gốc của kết nối để cho phép
	// các kết nối từ giao diện người dùng web của chúng ta.
	CheckOrigin: func(r *http.Request) bool {
		return true // Tạm thời cho phép tất cả các nguồn gốc
	},
}

// Client là một người dùng kết nối đến máy chủ.
type Client struct {
	// hub là trung tâm điều phối tin nhắn.
	hub *Hub

	// conn là kết nối WebSocket.
	conn *websocket.Conn

	// send là một kênh chứa các tin nhắn gửi đi.
	send chan []byte

	// userID là ID của người dùng
	userID string

	// conversationIDs là danh sách các cuộc trò chuyện mà user tham gia
	conversationIDs map[string]bool

	// lastActivity thời gian hoạt động cuối
	lastActivity time.Time
}

// Hub quản lý tất cả các client và tin nhắn.
type Hub struct {
	// clients là danh sách các client đã đăng ký.
	clients map[*Client]bool

	// broadcast là kênh nhận tin nhắn từ các client.
	broadcast chan []byte

	// register là kênh để đăng ký client mới.
	register chan *Client

	// unregister là kênh để hủy đăng ký client.
	unregister chan *Client

	// conversationClients map conversation ID -> danh sách clients
	conversationClients map[string]map[*Client]bool

	// userClients map user ID -> client
	userClients map[string]*Client
}

// newHub tạo một Hub mới.
func newHub() *Hub {
	return &Hub{
		broadcast:           make(chan []byte),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		clients:             make(map[*Client]bool),
		conversationClients: make(map[string]map[*Client]bool),
		userClients:         make(map[string]*Client),
	}
}

// run khởi chạy hub để xử lý các sự kiện.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.userClients[client.userID] = client

			// Gửi danh sách conversations hiện có cho client mới
			h.sendConversationList(client)

			log.Printf("Client đã kết nối: %s", client.userID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.userClients, client.userID)

				// Xóa khỏi tất cả conversations
				for convID := range client.conversationIDs {
					if clients, exists := h.conversationClients[convID]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.conversationClients, convID)
						}
					}
				}

				close(client.send)
				log.Printf("Client đã ngắt kết nối: %s", client.userID)
			}

		case message := <-h.broadcast:
			// Parse tin nhắn để xác định conversation
			var wsMsg models.WebSocketMessage
			if err := json.Unmarshal(message, &wsMsg); err == nil && wsMsg.ConvID != "" {
				// Gửi tin nhắn chỉ đến các client trong conversation
				if clients, exists := h.conversationClients[wsMsg.ConvID]; exists {
					for client := range clients {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client)
							delete(clients, client)
						}
					}
				}
			} else {
				// Broadcast đến tất cả clients (tin nhắn hệ thống)
				for client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

// JoinConversation thêm client vào conversation
func (h *Hub) JoinConversation(client *Client, conversationID string) {
	if h.conversationClients[conversationID] == nil {
		h.conversationClients[conversationID] = make(map[*Client]bool)
	}
	h.conversationClients[conversationID][client] = true
	client.conversationIDs[conversationID] = true

	// Gửi thông báo user joined đến các clients khác trong conversation
	joinMessage := models.WebSocketMessage{
		Type:   "user_joined",
		UserID: client.userID,
		ConvID: conversationID,
		Data:   client.userID,
	}

	if messageData, err := json.Marshal(joinMessage); err == nil {
		for otherClient := range h.conversationClients[conversationID] {
			if otherClient != client {
				select {
				case otherClient.send <- messageData:
				default:
					close(otherClient.send)
					delete(h.clients, otherClient)
					delete(h.conversationClients[conversationID], otherClient)
				}
			}
		}
	}

	log.Printf("Client %s đã tham gia conversation %s", client.userID, conversationID)
}

// LeaveConversation xóa client khỏi conversation
func (h *Hub) LeaveConversation(client *Client, conversationID string) {
	if clients, exists := h.conversationClients[conversationID]; exists {
		// Gửi thông báo user left đến các clients khác
		leaveMessage := models.WebSocketMessage{
			Type:   "user_left",
			UserID: client.userID,
			ConvID: conversationID,
			Data:   client.userID,
		}

		if messageData, err := json.Marshal(leaveMessage); err == nil {
			for otherClient := range clients {
				if otherClient != client {
					select {
					case otherClient.send <- messageData:
					default:
						close(otherClient.send)
						delete(h.clients, otherClient)
						delete(clients, otherClient)
					}
				}
			}
		}

		delete(clients, client)
		if len(clients) == 0 {
			delete(h.conversationClients, conversationID)
		}
	}
	delete(client.conversationIDs, conversationID)

	log.Printf("Client %s đã rời conversation %s", client.userID, conversationID)
}

// sendConversationList gửi danh sách conversations cho client
func (h *Hub) sendConversationList(client *Client) {
	// Tạo danh sách conversations demo
	conversations := map[string]interface{}{
		"general": map[string]interface{}{
			"id":           "general",
			"name":         "General Chat",
			"type":         "group",
			"participants": []string{"user1", "user2", "user3"},
			"lastMessage": map[string]interface{}{
				"content":   "Chào mọi người!",
				"timestamp": time.Now().Format(time.RFC3339),
				"sender":    "user1",
			},
		},
		"tech-talk": map[string]interface{}{
			"id":           "tech-talk",
			"name":         "Tech Talk",
			"type":         "group",
			"participants": []string{"dev1", "dev2"},
			"lastMessage": map[string]interface{}{
				"content":   "Go là ngôn ngữ tuyệt vời!",
				"timestamp": time.Now().Format(time.RFC3339),
				"sender":    "dev1",
			},
		},
	}

	conversationMessage := models.WebSocketMessage{
		Type: "conversation_list",
		Data: conversations,
	}

	if messageData, err := json.Marshal(conversationMessage); err == nil {
		select {
		case client.send <- messageData:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// CreateConversation tạo conversation mới
func (h *Hub) CreateConversation(client *Client, wsMsg models.WebSocketMessage) {
	// Parse conversation data từ message
	conversationData, ok := wsMsg.Data.(map[string]interface{})
	if !ok {
		log.Printf("Lỗi parse conversation data từ client %s", client.userID)
		return
	}

	name, nameOk := conversationData["name"].(string)
	convType, typeOk := conversationData["type"].(string)

	if !nameOk || !typeOk || name == "" {
		log.Printf("Thiếu thông tin conversation từ client %s", client.userID)
		return
	}

	// Tạo conversation ID unique
	conversationID := fmt.Sprintf("conv_%s_%d", client.userID, time.Now().Unix())

	// Tạo conversation object
	conversation := map[string]interface{}{
		"id":           conversationID,
		"name":         name,
		"type":         convType,
		"participants": []string{client.userID},
		"created_by":   client.userID,
		"created_at":   time.Now().Format(time.RFC3339),
		"lastMessage":  nil,
	}

	// Gửi thông báo conversation mới đến tất cả clients
	newConversationMessage := models.WebSocketMessage{
		Type:   "conversation_created",
		Data:   conversation,
		UserID: client.userID,
	}

	if messageData, err := json.Marshal(newConversationMessage); err == nil {
		// Broadcast đến tất cả clients
		for c := range h.clients {
			select {
			case c.send <- messageData:
			default:
				close(c.send)
				delete(h.clients, c)
			}
		}
	}

	log.Printf("Client %s đã tạo conversation mới: %s (%s)", client.userID, name, conversationID)
}

// readPump đọc tin nhắn từ kết nối WebSocket của client.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Parse tin nhắn WebSocket
		var wsMsg models.WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Lỗi parse tin nhắn: %v", err)
			continue
		}

		// Cập nhật thời gian hoạt động
		c.lastActivity = time.Now()

		// Xử lý các loại tin nhắn khác nhau
		switch wsMsg.Type {
		case "join_conversation":
			if convID, ok := wsMsg.Data.(string); ok {
				c.hub.JoinConversation(c, convID)
			}
		case "leave_conversation":
			if convID, ok := wsMsg.Data.(string); ok {
				c.hub.LeaveConversation(c, convID)
			}
		case "create_conversation":
			c.hub.CreateConversation(c, wsMsg)
		case "message", "typing", "reaction":
			// Gửi tin nhắn đến kênh broadcast
			wsMsg.UserID = c.userID // Đảm bảo tin nhắn có thông tin người gửi
			if updatedMessage, err := json.Marshal(wsMsg); err == nil {
				c.hub.broadcast <- updatedMessage
			} else {
				c.hub.broadcast <- message
			}
		default:
			// Gửi tin nhắn nhận được đến kênh broadcast của hub.
			c.hub.broadcast <- message
		}
	}
}

// writePump ghi tin nhắn từ kênh `send` của client vào kết nối WebSocket.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			// Hub đã đóng kênh.
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}
}

// serveWs xử lý các yêu cầu WebSocket từ client.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// Lấy userID từ query parameter hoặc header (tạm thời dùng cách đơn giản)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous" // Hoặc tạo một ID tạm thời
	}

	client := &Client{
		hub:             hub,
		conn:            conn,
		send:            make(chan []byte, 256),
		userID:          userID,
		conversationIDs: make(map[string]bool),
		lastActivity:    time.Now(),
	}
	client.hub.register <- client

	// Chạy goroutine để đọc và ghi tin nhắn đồng thời.
	go client.writePump()
	go client.readPump()
}

func main() {
	// Tạo một hub mới và chạy nó trong một goroutine.
	hub := newHub()
	go hub.run()

	// Cấu hình router HTTP.
	// Route "/" sẽ phục vụ tệp index.html.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../client/index.html")
	})

	// Route "/test" để test WebSocket cơ bản
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../test.html")
	})

	// Route "/debug" để debug WebSocket
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../debug.html")
	})

	// Route "/simple" để test WebSocket đơn giản
	http.HandleFunc("/simple", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../simple.html")
	})

	// Route "/ws" sẽ xử lý các kết nối WebSocket.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Khởi động máy chủ web.
	log.Println("Máy chủ đang chạy tại http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
