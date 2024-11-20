// controller/user_controller.go
package controllers

import (
	"context"
	"encoding/json"
	middleware "fire-watch/auth"
	"fire-watch/dbs"
	"fire-watch/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Hàm mã hóa mật khẩu
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Kiểm tra người dùng đã tồn tại hay chưa bằng email
func isUserExists(email string) (bool, error) {
	userCollection := models.GetUserCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Đăng ký người dùng mới
func RegisterUser(c *gin.Context) {
	userCollection := models.GetUserCollection()
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	// Validate dữ liệu server
	if err := user.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	// Kiểm tra người dùng đã tồn tại
	exists, err := isUserExists(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Mã hóa mật khẩu
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	// Gán ngày giờ hiện tại cho trường Createat
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Password = hashedPassword
	user.ID = primitive.NewObjectID() // Tạo ObjectID mới cho người dùng
	user.Status = 1                   // Mặc định trạng thái là hoạt động
	user.Role = "customer"            // Mặc định vai trò là customer

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Đăng nhập người dùng
func LoginUser(c *gin.Context) {
	userCollection := models.GetUserCollection()
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := userCollection.FindOne(ctx, bson.M{"email": credentials.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		}
		return
	}

	// Kiểm tra mật khẩu
	if !checkPasswordHash(credentials.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Tạo JWT token
	tokenString, err := middleware.CreateToken(user.ID.Hex(), user.Email, user.Username, user.Password, user.Role, user.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Lưu thông tin vào Redis
	userData := map[string]interface{}{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"role":    user.Role,
		"status":  user.Status,
	}

	userJSON, _ := json.Marshal(userData)
	if err := dbs.RedisClient.Set(ctx, "user:"+user.ID.Hex(), string(userJSON), 30*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data to Redis"})
		return
	}

	// Phản hồi đăng nhập thành công với token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user_id": user.ID.Hex(),
		"role":    user.Role,
	})
}

// Lấy tất cả người dùng
func GetAllUsers(c *gin.Context) {
	userCollection := models.GetUserCollection()
	var users []models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user"})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// Hiển thị người dùng theo ID
func GetUserByID(c *gin.Context) {
	userCollection := models.GetUserCollection()
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Cập nhật thông tin người dùng
func UpdateUser(c *gin.Context) {
	userCollection := models.GetUserCollection()
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}
	updatedUser.UpdatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"email":    updatedUser.Email,
			"username": updatedUser.Username,
			"status":   updatedUser.Status,
			"role":     updatedUser.Role,
		},
	}

	result, err := userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// Xóa người dùng
func DeleteUser(c *gin.Context) {
	userCollection := models.GetUserCollection()
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = userCollection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Kiểm tra mật khẩu
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
