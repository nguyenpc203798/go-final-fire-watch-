// controllers/category_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Thêm category mới
func AddCategory(c *gin.Context) {
	categoryCollection := models.GetCategoryCollection()
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category data"})
		return
	}

	category.ID = primitive.NewObjectID()
	category.Status = 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := categoryCollection.InsertOne(ctx, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting category"})
		return
	}

	// Xóa cache Redis sau khi thêm category mới
	dbs.RedisClient.Del(ctx, "categories")

	c.JSON(http.StatusOK, category)
}

// Lấy tất cả categories
func GetAllCategories(c *gin.Context) {
	categoryCollection := models.GetCategoryCollection()
	var categories []models.Category

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedCategories, err := dbs.RedisClient.Get(ctx, "categories").Result()
	if err == nil && cachedCategories != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedCategories), &categories)
		c.JSON(http.StatusOK, categories)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := categoryCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var category models.Category
		if err := cursor.Decode(&category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding category"})
			return
		}
		categories = append(categories, category)
	}

	// Lưu dữ liệu vào Redis cache
	categoriesJSON, _ := json.Marshal(categories)
	dbs.RedisClient.Set(ctx, "categories", string(categoriesJSON), 30*time.Minute)

	c.JSON(http.StatusOK, categories)
}

// Lấy một category theo ID
func GetCategoryByID(c *gin.Context) {
	categoryCollection := models.GetCategoryCollection()
	id := c.Query("id")
	categoryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachedCategory, err := dbs.RedisClient.Get(ctx, "category_"+id).Result()
	if err == nil && cachedCategory != "" {
		// Nếu có cache, trả về cache
		var category models.Category
		json.Unmarshal([]byte(cachedCategory), &category)
		c.JSON(http.StatusOK, category)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var category models.Category
	err = categoryCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching category"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	categoryJSON, _ := json.Marshal(category)
	dbs.RedisClient.Set(ctx, "category_"+id, string(categoryJSON), 30*time.Minute)

	c.JSON(http.StatusOK, category)
}

// Cập nhật category
func UpdateCategory(c *gin.Context) {
	categoryCollection := models.GetCategoryCollection()
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": category.ID}
	update := bson.M{"$set": category}

	_, err := categoryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "category_"+category.ID.Hex())
	dbs.RedisClient.Del(ctx, "categories")

	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully"})
}

// DeleteCategory xóa một category và tự động cập nhật các bộ phim có chứa category này
func DeleteCategory(c *gin.Context) {
	categoryCollection := models.GetCategoryCollection()
	movieCollection := models.GetMovieCollection()
	id := c.Param("id")
	categoryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Xóa category trong categoryCollection
	_, err = categoryCollection.DeleteOne(ctx, bson.M{"_id": categoryID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting category"})
		return
	}

	// Cập nhật tất cả các bộ phim, xóa categoryID khỏi mảng Category của Movie
	_, err = movieCollection.UpdateMany(
		ctx,
		bson.M{"category": categoryID}, // Tìm những bộ phim chứa categoryID
		bson.M{"$pull": bson.M{"category": categoryID}}, // Xóa categoryID khỏi mảng category
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movies"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "category_"+id)
	dbs.RedisClient.Del(ctx, "categories")
	dbs.RedisClient.Del(ctx, "movies_cache")

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted and movies updated successfully"})
}
