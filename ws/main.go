package main

import (
	"log"
	"net/http"

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
}

// newHub tạo một Hub mới.
func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// run khởi chạy hub để xử lý các sự kiện.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
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
		// Gửi tin nhắn nhận được đến kênh broadcast của hub.
		c.hub.broadcast <- message
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
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
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
		http.ServeFile(w, r, "index.html")
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
