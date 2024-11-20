// models/quality.go
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

type Quality struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" form:"id"`
	MovieID     primitive.ObjectID `bson:"movie_id" form:"movie_ids" validate:"required"`
	EpisodeID   primitive.ObjectID `bson:"episode_id" form:"episode_ids" validate:"required"`
	ServerID    primitive.ObjectID `bson:"server_id" form:"server_id" validate:"required"`
	Title       string             `bson:"title" form:"title" validate:"required,min=1,max=100"`
	Description string             `bson:"description" form:"description" validate:"omitempty,max=250"`
	Videourl    string             `bson:"videourl" form:"videourl" validate:"required,min=3,max=100"`
	Status      int                `bson:"status" form:"status"`
	Deleted     string             `bson:"deleted, omitempty" form:"deleted"`
	CreatedAt   time.Time          `bson:"created_at" form:"created_at"` // Sửa lại tên trường ở đây
	UpdatedAt   time.Time          `bson:"updated_at" form:"updated_at"` // Tương tự với updated_at
}

// Khai báo biến collection cho quality
var qualityCollection *mongo.Collection

// Khởi tạo qualityCollection
func InitializeQualityCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	qualityCollection = dbs.DB.Collection("qualities")
}

// Hàm này trả về collection của Quality để controller có thể sử dụng lại
func GetQualityCollection() *mongo.Collection {
	return qualityCollection
}

// Khởi tạo validator
var validatequality = validator.New()

// Validate method for Quality struct
func (quality *Quality) Validate() error {
	validate := validator.New()

	// Validate struct fields
	if err := validate.Struct(quality); err != nil {
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
				case "dive":
					errorMessages = append(errorMessages, fieldErr.Field()+" contains invalid elements")
				default:
					errorMessages = append(errorMessages, fieldErr.Field()+" is invalid")
				}
			}
			// Trả về một lỗi tổng hợp từ các thông báo lỗi chi tiết
			return errors.New("Validation failed: " + joinErrorsQuality(errorMessages))
		}
		return err
	}
	return nil
}

// Hàm joinErrors để nối các thông báo lỗi thành một chuỗi
func joinErrorsQuality(errors []string) string {
	return strings.Join(errors, ", ")
}
