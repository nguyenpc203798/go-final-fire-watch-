// controllers/category_controller.go
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

// Thêm category mới
func AddCategory(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo biến chứa dữ liệu từ form
	categoryCollection := models.GetCategoryCollection()
	var category models.Category

	// Bind dữ liệu từ form (multipart/form-data hoặc application/x-www-form-urlencoded)
	if err := c.ShouldBind(&category); err != nil {
		// Trả về thông báo lỗi validate dữ liệu
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu category
	if err := category.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(), // Trả về danh sách các lỗi chi tiết
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm category với tiêu đề trùng lặp
	var existingCategory models.Category
	if err := categoryCollection.FindOne(ctx, bson.M{"title": category.Title}).Decode(&existingCategory); err == nil {
		// Nếu tìm thấy, trả về thông báo lỗi
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Category already exists",
			"message": "A category with this title already exists. Please choose a different title.",
		})
		return
	}

	// Gán giá trị ID và trạng thái mặc định
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	// Thực hiện thêm category mới vào MongoDB
	_, err := categoryCollection.InsertOne(ctx, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add category",
			"message": "Unable to add new category due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "categories")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new category was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Category added successfully!",
	})
}

// Lấy tất cả categories
func GetAllCategories() ([]models.Category, error) {
	categoryCollection := models.GetCategoryCollection()
	var categories []models.Category

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedCategories, err := dbs.RedisClient.Get(ctx, "categories").Result()
	if err == nil && cachedCategories != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedCategories), &categories)
		return categories, nil
	}

	// Điều kiện lọc: chỉ lấy những category chưa bị xóa (Deleted != 1)
	filter := bson.M{
		"$or": []bson.M{
			{"deleted": bson.M{"$ne": "deleted"}}, // Category có trường Deleted khác deleted
		},
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := categoryCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var category models.Category
		if err := cursor.Decode(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	// Lưu dữ liệu vào Redis cache để tránh phải truy vấn lại
	categoriesJSON, _ := json.Marshal(categories)
	dbs.RedisClient.Set(ctx, "categories", string(categoriesJSON), 30*time.Minute)

	return categories, nil
}

// UpdateCategory cập nhật thông tin của một category
func UpdateCategory(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy collection category
	categoryCollection := models.GetCategoryCollection()
	var category models.Category

	// Lấy ID từ URL và kiểm tra ID hợp lệ hay không
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided category ID is not valid",
		})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON data",
			"message": err.Error(),
		})
		return
	}

	// Validate dữ liệu category
	if err := category.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}
	category.UpdatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm và kiểm tra xem category có tồn tại không
	var existingCategory models.Category
	err = categoryCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingCategory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Nếu không tìm thấy category
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Category not found",
				"message": "No category found with the provided ID",
			})
		} else {
			// Nếu có lỗi truy vấn
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to find category",
				"message": "Unable to find category due to a server error. Please try again later.",
			})
		}
		return
	}

	// Cập nhật các trường của category
	update := bson.M{
		"$set": bson.M{
			"title":       category.Title,
			"description": category.Description,
			"status":      category.Status,
			"slug":        category.Slug,
		},
	}

	// Thực hiện cập nhật category trong MongoDB
	_, err = categoryCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update category",
			"message": "Unable to update category due to a server error. Please try again later.",
		})
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "categories")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new category was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully!",
	})
}

// DeleteCategory là hàm xử lý yêu cầu xóa ảo
func DeleteCategory(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	categoryCollection := models.GetCategoryCollection()
	// Lấy ID từ route
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Tạo filter để tìm Category theo ID
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

	_, err = categoryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa danh mục"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "categories")

	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new category was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Danh mục đã được xóa"})
}

// Hàm UpdateCategoryField để cập nhật trường cụ thể của Category
func UpdateCategoryField(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của Category từ tham số URL
	idParam := c.Param("id")
	categoryID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
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

	collection := models.GetCategoryCollection()
	filter := bson.M{"_id": categoryID}
	_, err = collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category field"})
		return
	}
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "categories")
	// Gửi thông điệp tới tất cả các client qua WebSocket
	message := []byte("A new category was updated!")
	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(message))

	websocketServer.BroadcastMessage(message)
	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category field updated successfully",
	})
}
