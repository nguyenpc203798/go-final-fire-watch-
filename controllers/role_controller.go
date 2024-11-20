// controllers/role_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

// Khai báo biến collection cho role
var roleCollection *mongo.Collection

// Khởi tạo roleCollection
func InitializeroleCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	roleCollection = dbs.DB.Collection("roles")
}

// Thêm role mới
func AddRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
		return
	}

	role.ID = primitive.NewObjectID()
	role.Status = 1 

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := roleCollection.InsertOne(ctx, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting role"})
		return
	}

	// Xóa cache Redis sau khi thêm role mới
	dbs.RedisClient.Del(ctx, "roles")

	c.JSON(http.StatusOK, role)
}

// Lấy tất cả roles
func GetAllRoles(c *gin.Context) {
	var roles []models.Role

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedroles, err := dbs.RedisClient.Get(ctx, "roles").Result()
	if err == nil && cachedroles != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedroles), &roles)
		c.JSON(http.StatusOK, roles)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := roleCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching roles"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var role models.Role
		if err := cursor.Decode(&role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding role"})
			return
		}
		roles = append(roles, role)
	}

	// Lưu dữ liệu vào Redis cache
	rolesJSON, _ := json.Marshal(roles)
	dbs.RedisClient.Set(ctx, "roles", string(rolesJSON), 30*time.Minute)

	c.JSON(http.StatusOK, roles)
}

// Lấy một role theo ID
func GetRoleByID(c *gin.Context) {
	id := c.Query("id")
	roleID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachedrole, err := dbs.RedisClient.Get(ctx, "role_"+id).Result()
	if err == nil && cachedrole != "" {
		// Nếu có cache, trả về cache
		var role models.Role
		json.Unmarshal([]byte(cachedrole), &role)
		c.JSON(http.StatusOK, role)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var role models.Role
	err = roleCollection.FindOne(ctx, bson.M{"_id": roleID}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching role"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	roleJSON, _ := json.Marshal(role)
	dbs.RedisClient.Set(ctx, "role_"+id, string(roleJSON), 30*time.Minute)

	c.JSON(http.StatusOK, role)
}

// Cập nhật role
func UpdateRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": role.ID}
	update := bson.M{"$set": role}

	_, err := roleCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating role"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "role_"+role.ID.Hex())
	dbs.RedisClient.Del(ctx, "roles")

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

// Xóa role
func DeleteRole(c *gin.Context) {
	id := c.Query("id")
	roleID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = roleCollection.DeleteOne(ctx, bson.M{"_id": roleID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting role"})
		return
	}

	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "role_"+id)
	dbs.RedisClient.Del(ctx, "roles")

	c.JSON(http.StatusOK, gin.H{"message": "role deleted successfully"})
}
