// models/episode.go
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

type Episode struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty"`
	MovieID       primitive.ObjectID   `bson:"movieid" form:"movieid" validate:"required"`
	Number        int                  `bson:"number" form:"number" validate:"required"`
	Image         string               `bson:"image,omitempty" form:"image"`
	Server        []primitive.ObjectID `bson:"server,omitempty" form:"server"`
	Status        int                  `bson:"status" form:"status" validate:"required"`
	Deleted       string               `bson:"deleted,omitempty" form:"deleted"`
	CreatedAt     time.Time            `bson:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at"`
	ServerDetails []Server             `bson:"serverDetails,omitempty"` // Chắc chắn là mảng
}

// Khai báo biến collection cho episode
var episodeCollection *mongo.Collection

// Khởi tạo episodeCollection
func InitializeEpisodeCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	episodeCollection = dbs.DB.Collection("episodes")
}

// Hàm này trả về collection của Episode để controller có thể sử dụng lại
func GetEpisodeCollection() *mongo.Collection {
	return episodeCollection
}

// Khởi tạo validator
var validateepisode = validator.New()

// Validate method for Episode struct
func (episode *Episode) Validate() error {
	validate := validator.New()

	// Validate struct fields
	if err := validate.Struct(episode); err != nil {
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
			return errors.New("Validation failed: " + joinErrorsEpisode(errorMessages))
		}
		return err
	}
	return nil
}

// Hàm joinErrors để nối các thông báo lỗi thành một chuỗi
func joinErrorsEpisode(errors []string) string {
	return strings.Join(errors, ", ")
}
