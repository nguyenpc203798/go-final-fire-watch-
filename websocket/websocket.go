package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client đại diện cho một kết nối WebSocket
type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

// WebSocketServer quản lý tất cả các kết nối WebSocket
type WebSocketServer struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.Mutex
}

// NewWebSocketServer tạo một WebSocket server mới
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run bắt đầu server WebSocket và lắng nghe các sự kiện
func (server *WebSocketServer) Run() {
	for {
		select {
		case client := <-server.Register:
			server.Mutex.Lock()
			server.Clients[client] = true
			server.Mutex.Unlock()
			log.Println("Client connected")

		case client := <-server.Unregister:
			server.Mutex.Lock()
			if _, ok := server.Clients[client]; ok {
				delete(server.Clients, client)
				close(client.Send)
				log.Println("Client disconnected")
			}
			server.Mutex.Unlock()

		case message := <-server.Broadcast:
			server.Mutex.Lock()
			for client := range server.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(server.Clients, client)
				}
			}
			server.Mutex.Unlock()
		}
	}
}

// HandleConnections xử lý yêu cầu kết nối WebSocket
func (server *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Cho phép kết nối từ bất kỳ domain nào
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	client := &Client{Conn: conn, Send: make(chan []byte)}

	// Đăng ký client mới
	server.Register <- client

	// Đọc tin nhắn từ client và gửi vào channel Broadcast
	go server.handleMessages(client)

	// Gửi tin nhắn từ channel Send về lại client
	go server.sendMessages(client)
}

// handleMessages xử lý tin nhắn đến từ client
func (server *WebSocketServer) handleMessages(client *Client) {
	defer func() {
		server.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected WebSocket closure: %v", err)
			} else {
				log.Println("Client disconnected:", err)
			}
			break
		}

		// Gửi tin nhắn tới tất cả các client
		server.Broadcast <- message
	}
}
// Gửi tin nhắn tới tất cả client khi có sự kiện xảy ra
func (server *WebSocketServer) BroadcastMessage(message []byte) {
    server.Broadcast <- message // Đẩy tin nhắn vào channel Broadcast
}

// sendMessages gửi tin nhắn từ server đến client
func (server *WebSocketServer) sendMessages(client *Client) {
	defer client.Conn.Close()
	for message := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Error sending message:", err)
			break
		}
	}
}
