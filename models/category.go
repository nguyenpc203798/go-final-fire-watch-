// models/category.go
package models

import (
	"errors"
	"fire-watch/dbs" // Điều chỉnh đường dẫn tùy thuộc vào cấu trúc dự án của bạn
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10" // Thêm validator
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Category struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" form:"id"`
	Title       string             `bson:"title" form:"title" validate:"required,min=3,max=100"`
	Description string             `bson:"description" form:"description" validate:"omitempty,max=250"`
	Status      int                `bson:"status" form:"status"`
	Slug        string             `bson:"slug" form:"slug" validate:"required"`
	Deleted     string             `bson:"deleted, omitempty" form:"deleted"`
	CreatedAt   time.Time          `bson:"created_at" form:"created_at"` // Sửa lại tên trường ở đây
	UpdatedAt   time.Time          `bson:"updated_at" form:"updated_at"` // Tương tự với updated_at
}

// Khai báo biến collection cho category
var categoryCollection *mongo.Collection

// Khởi tạo categoryCollection
func InitializeCategoryCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	categoryCollection = dbs.DB.Collection("categories")
}

// Hàm này trả về collection của Category để controller có thể sử dụng lại
func GetCategoryCollection() *mongo.Collection {
	return categoryCollection
}

// Khởi tạo validator
var validatecategory = validator.New()

// Validate method for Category struct
func (category *Category) Validate() error {
	validate := validator.New()

	// Validate struct fields
	if err := validate.Struct(category); err != nil {
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
				case "oneof":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be either "+fieldErr.Param())
				default:
					errorMessages = append(errorMessages, fieldErr.Field()+" is invalid")
				}
			}
			// Trả về một lỗi tổng hợp từ các thông báo lỗi chi tiết
			return errors.New("Validation failed: " + joinErrorsCategory(errorMessages))
		}
		return err
	}
	return nil
}

// Hàm joinErrors để nối các thông báo lỗi thành một chuỗi
func joinErrorsCategory(errors []string) string {
	return strings.Join(errors, ", ")
}
