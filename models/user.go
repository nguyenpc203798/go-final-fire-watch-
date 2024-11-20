package models

import (
	"errors"
	"fire-watch/dbs"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

	// Định nghĩa struct User
	type User struct {
		ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`                                                     // ID của người dùng
		Username  string             `bson:"username"  form:"username"  json:"username" validate:"required,min=3,max=50"` // Tên đăng nhập, yêu cầu từ 3 đến 50 ký tự
		Email     string             `bson:"email" form:"email" json:"email" validate:"required,email"`                   // Địa chỉ email, yêu cầu định dạng email hợp lệ
		Password  string             `bson:"password" form:"password" json:"password" validate:"required,min=6"`          // Mật khẩu, yêu cầu tối thiểu 6 ký tự
		Role      string             `bson:"role" form:"role" json:"role"`                                                // Role phải là 'admin' hoặc 'user'
		Status    int                `bson:"status" form:"status" json:"status"`                                          // Trạng thái (1: hoạt động, 0: không hoạt động)
		Deleted   string             `bson:"deleted, omitempty" form:"deleted"`
		CreatedAt time.Time          `bson:"create_at"` // Thời gian tạo
		UpdatedAt time.Time          `bson:"update_at"` // Thời gian cập nhật
	}

// Khai báo biến collection cho user
var userCollection *mongo.Collection

// Khởi tạo userCollection
func InitializeUserCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	userCollection = dbs.DB.Collection("users")
}

// Hàm này trả về collection của User để controller có thể sử dụng lại
func GetUserCollection() *mongo.Collection {
	return userCollection
}

// Khởi tạo validator
var validateuser = validator.New()

// Validate method for User struct
func (user *User) Validate() error {
	validate := validator.New()

	// Validate struct fields
	if err := validate.Struct(user); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Tạo một slice chứa thông báo lỗi chi tiết
			var errorMessages []string
			for _, fieldErr := range validationErrors {
				// Xử lý thông báo lỗi chi tiết dựa trên trường và loại lỗi
				switch fieldErr.Tag() {
				case "required":
					errorMessages = append(errorMessages, fieldErr.Field()+" is required")
				case "min":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be at least "+fieldErr.Param()+" characters")
				case "max":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be less than "+fieldErr.Param()+" characters")
				case "email":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be a valid email address")
				case "oneof":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be one of: "+fieldErr.Param())
				default:
					errorMessages = append(errorMessages, fieldErr.Field()+" is invalid")
				}
			}
			// Trả về một lỗi tổng hợp từ các thông báo lỗi chi tiết
			return errors.New("Validation failed: " + joinErrorsUser(errorMessages))
		}
		return err
	}
	return nil
}

// Hàm joinErrors để nối các thông báo lỗi thành một chuỗi
func joinErrorsUser(errors []string) string {
	return strings.Join(errors, ", ")
}
