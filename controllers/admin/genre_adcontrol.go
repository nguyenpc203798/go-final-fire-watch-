// controllers/genre_controller.go
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

// Thêm genre mới
func AddGenre(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo biến chứa dữ liệu từ form
	genreCollection := models.GetGenreCollection()
	var genre models.Genre

	// Bind dữ liệu từ form (multipart/form-data hoặc application/x-www-form-urlencoded)
	if err := c.ShouldBind(&genre); err != nil {
		// Trả về thông báo lỗi validate dữ liệu
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu genre
	if err := genre.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(), // Trả về danh sách các lỗi chi tiết
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm genre với tiêu đề trùng lặp
	var existingGenre models.Genre
	if err := genreCollection.FindOne(ctx, bson.M{"title": genre.Title}).Decode(&existingGenre); err == nil {
		// Nếu tìm thấy, trả về thông báo lỗi
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Genre already exists",
			"message": "A genre with this title already exists. Please choose a different title.",
		})
		return
	}

	// Gán giá trị ID và trạng thái mặc định
	genre.ID = primitive.NewObjectID()
	genre.CreatedAt = time.Now()
	genre.UpdatedAt = time.Now()

	// Thực hiện thêm genre mới vào MongoDB
	_, err := genreCollection.InsertOne(ctx, genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add genre",
			"message": "Unable to add new genre due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "genres")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new genre was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Genre added successfully!",
	})
}

// Lấy tất cả genres
func GetAllGenres() ([]models.Genre, error) {
	genreCollection := models.GetGenreCollection()
	var genres []models.Genre

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedGenres, err := dbs.RedisClient.Get(ctx, "genres").Result()
	if err == nil && cachedGenres != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedGenres), &genres)
		return genres, nil
	}

	// Điều kiện lọc: chỉ lấy những genre chưa bị xóa (Deleted != 1)
	filter := bson.M{
		"$or": []bson.M{
			{"deleted": bson.M{"$ne": "deleted"}}, // Genre có trường Deleted khác deleted
		},
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := genreCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var genre models.Genre
		if err := cursor.Decode(&genre); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	// Lưu dữ liệu vào Redis cache để tránh phải truy vấn lại
	genresJSON, _ := json.Marshal(genres)
	dbs.RedisClient.Set(ctx, "genres", string(genresJSON), 30*time.Minute)

	return genres, nil
}

// UpdateGenre cập nhật thông tin của một genre
func UpdateGenre(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy collection genre
	genreCollection := models.GetGenreCollection()
	var genre models.Genre

	// Lấy ID từ URL và kiểm tra ID hợp lệ hay không
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided genre ID is not valid",
		})
		return
	}

	if err := c.ShouldBindJSON(&genre); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu genre
	if err := genre.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}
	genre.UpdatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm và kiểm tra xem genre có tồn tại không
	var existingGenre models.Genre
	err = genreCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingGenre)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Nếu không tìm thấy genre
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Genre not found",
				"message": "No genre found with the provided ID",
			})
		} else {
			// Nếu có lỗi truy vấn
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to find genre",
				"message": "Unable to find genre due to a server error. Please try again later.",
			})
		}
		return
	}

	// Cập nhật các trường của genre
	update := bson.M{
		"$set": bson.M{
			"title":       genre.Title,
			"description": genre.Description,
			"status":      genre.Status,
			"slug":        genre.Slug,
		},
	}

	// Thực hiện cập nhật genre trong MongoDB
	_, err = genreCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update genre",
			"message": "Unable to update genre due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "genres")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new genre was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Genre updated successfully!",
	})
}

// DeleteGenre là hàm xử lý yêu cầu xóa ảo
func DeleteGenre(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	genreCollection := models.GetGenreCollection()
	// Lấy ID từ route
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Tạo filter để tìm Genre theo ID
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

	_, err = genreCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa danh mục"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "genres")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new genre was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Danh mục đã được xóa"})
}

// Hàm UpdateGenreField để cập nhật trường cụ thể của Genre
func UpdateGenreField(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của Genre từ tham số URL
	idParam := c.Param("id")
	genreID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre ID"})
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

	collection := models.GetGenreCollection()
	filter := bson.M{"_id": genreID}
	_, err = collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update genre field"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "genres")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new genre was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Genre field updated successfully",
	})
}
