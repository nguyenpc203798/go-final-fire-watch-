// controllers/server_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fire-watch/websocket"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Thêm server mới
func AddServer(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo biến chứa dữ liệu từ form
	serverCollection := models.GetServerCollection()
	var server models.Server

	// Bind dữ liệu từ form (multipart/form-data hoặc application/x-www-form-urlencoded)
	if err := c.ShouldBind(&server); err != nil {
		// Trả về thông báo lỗi validate dữ liệu
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu server
	if err := server.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(), // Trả về danh sách các lỗi chi tiết
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm server với tiêu đề trùng lặp
	var existingServer models.Server
	if err := serverCollection.FindOne(ctx, bson.M{"title": server.Title}).Decode(&existingServer); err == nil {
		// Nếu tìm thấy, trả về thông báo lỗi
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Server already exists",
			"message": "A server with this title already exists. Please choose a different title.",
		})
		return
	}

	// Gán giá trị ID và trạng thái mặc định
	server.ID = primitive.NewObjectID()
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()
	server.MovieIDs = make([]primitive.ObjectID, 0) // Luôn khởi tạo là mảng rỗng
	server.EpisodeIDs = make([]primitive.ObjectID, 0) // Luôn khởi tạo là mảng rỗng
	server.Quality = make([]primitive.ObjectID, 0) // Luôn khởi tạo là mảng rỗng

	// Thực hiện thêm server mới vào MongoDB
	_, err := serverCollection.InsertOne(ctx, server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add server",
			"message": "Unable to add new server due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "servers")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new server was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Server added successfully!",
	})
}

// Lấy tất cả servers
func GetAllServers() ([]models.Server, error) {
	serverCollection := models.GetServerCollection()
	var servers []models.Server

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedServers, err := dbs.RedisClient.Get(ctx, "servers").Result()
	if err == nil && cachedServers != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedServers), &servers)
		return servers, nil
	}

	// Điều kiện lọc: chỉ lấy những server chưa bị xóa (Deleted != 1)
	filter := bson.M{
		"$or": []bson.M{
			{"deleted": bson.M{"$ne": "deleted"}}, // Server có trường Deleted khác deleted
		},
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := serverCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var server models.Server
		if err := cursor.Decode(&server); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	// Lưu dữ liệu vào Redis cache để tránh phải truy vấn lại
	serversJSON, _ := json.Marshal(servers)
	dbs.RedisClient.Set(ctx, "servers", string(serversJSON), 30*time.Minute)

	return servers, nil
}

// UpdateServer cập nhật thông tin của một server
func UpdateServer(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy collection server
	serverCollection := models.GetServerCollection()
	var server models.Server

	// Lấy ID từ URL và kiểm tra ID hợp lệ hay không
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided server ID is not valid",
		})
		return
	}

	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu server
	if err := server.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}
	server.UpdatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm và kiểm tra xem server có tồn tại không
	var existingServer models.Server
	err = serverCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingServer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Nếu không tìm thấy server
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Server not found",
				"message": "No server found with the provided ID",
			})
		} else {
			// Nếu có lỗi truy vấn
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to find server",
				"message": "Unable to find server due to a server error. Please try again later.",
			})
		}
		return
	}

	// Cập nhật các trường của server
	update := bson.M{
		"$set": bson.M{
			"title":       server.Title,
			"description": server.Description,
			"status":      server.Status,
			"slug":        server.Slug,
		},
	}

	// Thực hiện cập nhật server trong MongoDB
	_, err = serverCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update server",
			"message": "Unable to update server due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "servers")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new server was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Server updated successfully!",
	})
}

// DeleteServer là hàm xử lý yêu cầu xóa ảo
func DeleteServer(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	serverCollection := models.GetServerCollection()
	// Lấy ID từ route
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Tạo filter để tìm Server theo ID
	filter := bson.M{"_id": objectID}

	// Tạo update để cập nhật trường Deleted = 1
	update := bson.M{
		"$set": bson.M{
			"deleted": "deleted",
		},
	}

	// Cập nhật trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = serverCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa danh mục"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "servers")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new server was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Danh mục đã được xóa"})
}

// Hàm UpdateServerField để cập nhật trường cụ thể của Server
func UpdateServerField(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của Server từ tham số URL
	idParam := c.Param("id")
	serverID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	// Nhận dữ liệu từ request body
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Kiểm tra xem request có chứa trường "field" và "value" hay không
	field, fieldOk := requestData["field"].(string)
	value, valueOk := requestData["value"]
	if !fieldOk || !valueOk {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'field' or 'value'"})
		return
	}

	// Tạo điều kiện cập nhật trường tương ứng trong MongoDB
	updateData := bson.M{
		"$set": bson.M{
			field: value,
		},
	}

	// Cập nhật trường trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := models.GetServerCollection()
	filter := bson.M{"_id": serverID}
	_, err = collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server field"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "servers")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new server was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Server field updated successfully",
	})
}
