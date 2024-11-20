// controllers/country_controller.go
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

// Thêm country mới
func AddCountry(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo biến chứa dữ liệu từ form
	countryCollection := models.GetCountryCollection()
	var country models.Country

	// Bind dữ liệu từ form (multipart/form-data hoặc application/x-www-form-urlencoded)
	if err := c.ShouldBind(&country); err != nil {
		// Trả về thông báo lỗi validate dữ liệu
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu country
	if err := country.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(), // Trả về danh sách các lỗi chi tiết
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm country với tiêu đề trùng lặp
	var existingCountry models.Country
	if err := countryCollection.FindOne(ctx, bson.M{"title": country.Title}).Decode(&existingCountry); err == nil {
		// Nếu tìm thấy, trả về thông báo lỗi
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Country already exists",
			"message": "A country with this title already exists. Please choose a different title.",
		})
		return
	}

	// Gán giá trị ID và trạng thái mặc định
	country.ID = primitive.NewObjectID()
	country.CreatedAt = time.Now()
	country.UpdatedAt = time.Now()

	// Thực hiện thêm country mới vào MongoDB
	_, err := countryCollection.InsertOne(ctx, country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add country",
			"message": "Unable to add new country due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "countries")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new country was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Country added successfully!",
	})
}

// Lấy tất cả countries
func GetAllCountries() ([]models.Country, error) {
	countryCollection := models.GetCountryCollection()
	var countries []models.Country

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedCountries, err := dbs.RedisClient.Get(ctx, "countries").Result()
	if err == nil && cachedCountries != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedCountries), &countries)
		return countries, nil
	}

	// Điều kiện lọc: chỉ lấy những country chưa bị xóa (Deleted != 1)
	filter := bson.M{
		"$or": []bson.M{
			{"deleted": bson.M{"$ne": "deleted"}}, // Country có trường Deleted khác deleted
		},
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := countryCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var country models.Country
		if err := cursor.Decode(&country); err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}

	// Lưu dữ liệu vào Redis cache để tránh phải truy vấn lại
	countriesJSON, _ := json.Marshal(countries)
	dbs.RedisClient.Set(ctx, "countries", string(countriesJSON), 30*time.Minute)

	return countries, nil
}

// UpdateCountry cập nhật thông tin của một country
func UpdateCountry(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy collection country
	countryCollection := models.GetCountryCollection()
	var country models.Country

	// Lấy ID từ URL và kiểm tra ID hợp lệ hay không
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided country ID is not valid",
		})
		return
	}

	if err := c.ShouldBindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu country
	if err := country.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}
	country.UpdatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm và kiểm tra xem country có tồn tại không
	var existingCountry models.Country
	err = countryCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingCountry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Nếu không tìm thấy country
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Country not found",
				"message": "No country found with the provided ID",
			})
		} else {
			// Nếu có lỗi truy vấn
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to find country",
				"message": "Unable to find country due to a server error. Please try again later.",
			})
		}
		return
	}

	// Cập nhật các trường của country
	update := bson.M{
		"$set": bson.M{
			"title":       country.Title,
			"description": country.Description,
			"status":      country.Status,
			"slug":        country.Slug,
		},
	}

	// Thực hiện cập nhật country trong MongoDB
	_, err = countryCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update country",
			"message": "Unable to update country due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "countries")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new country was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Country updated successfully!",
	})
}

// DeleteCountry là hàm xử lý yêu cầu xóa ảo
func DeleteCountry(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	countryCollection := models.GetCountryCollection()
	// Lấy ID từ route
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Tạo filter để tìm Country theo ID
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

	_, err = countryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa danh mục"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "countries")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new country was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Danh mục đã được xóa"})
}

// Hàm UpdateCountryField để cập nhật trường cụ thể của Country
func UpdateCountryField(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của Country từ tham số URL
	idParam := c.Param("id")
	countryID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country ID"})
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

	collection := models.GetCountryCollection()
	filter := bson.M{"_id": countryID}
	_, err = collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update country field"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "countries")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new country was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Country field updated successfully",
	})
}
