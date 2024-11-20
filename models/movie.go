// models/movie.go
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

// Movie struct mô phỏng bảng movies
type Movie struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" form:"id"`
	Title           string               `bson:"title" form:"title" validate:"required,min=3,max=100"`
	NameEng         string               `bson:"name_eng" form:"name_eng" validate:"omitempty,max=100"`
	Description     string               `bson:"description" form:"description" validate:"omitempty"`
	Tags            string               `bson:"tags,omitempty" form:"tags"`
	Status          int                  `bson:"status" form:"status" validate:"required"`
	Image           string               `bson:"image,omitempty" form:"image"`
	Moreimage       []string             `bson:"moreimage,omitempty" form:"moreimage" validate:"omitempty,dive"`
	Slug            string               `bson:"slug" form:"slug" validate:"required"`
	Category        []primitive.ObjectID `bson:"category" form:"category" validate:"required,dive"`
	Genre           []primitive.ObjectID `bson:"genre" form:"genre" validate:"required,dive"`
	Country         primitive.ObjectID   `bson:"country,omitempty" form:"country" validate:"omitempty"`
	Episode         []primitive.ObjectID `bson:"episode" form:"episode"`
	Hotmovie        int                  `bson:"hotmovie" form:"hotmovie" validate:"omitempty,oneof=1 2"`
	MaxQuality      int                  `bson:"maxquality,omitempty" form:"maxquality" validate:"omitempty,oneof=1 720 1080 1440 2160"`
	Sub             []string             `bson:"sub,omitempty" form:"sub" validate:"omitempty"`
	Trailer         string               `bson:"trailer,omitempty" form:"trailer" validate:"omitempty"`
	Year            int                  `bson:"year,omitempty" form:"year" validate:"omitempty,numeric"`
	Season          int                  `bson:"season,omitempty" form:"season" validate:"omitempty"`
	Duration        string               `bson:"duration,omitempty" form:"duration"`
	Rating          int                  `bson:"rating,omitempty" form:"rating"`
	Comment         string               `bson:"comment,omitempty" form:"comment"`
	Numofep         int                  `bson:"numofep,omitempty" form:"numofep" validate:"omitempty"`
	Views           int                  `bson:"views,omitempty" form:"views" validate:"omitempty"`
	Position        int                  `bson:"position,omitempty" form:"position"` // Thêm trường position
	CreatedAt       time.Time            `bson:"created_at" form:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" form:"updated_at"`
	Deleted         string               `bson:"deleted, omitempty" form:"deleted"`
	CategoryDetails []Category           `bson:"categoryDetails,omitempty"`
	GenreDetails    []Genre              `bson:"genreDetails,omitempty"`
	CountryDetails  []Country            `bson:"countryDetails,omitempty"`
	EpisodeDetails  []Episode            `bson:"episodeDetails,omitempty"`
}

// Khai báo biến collection cho movie
var movieCollection *mongo.Collection

// Khởi tạo movieCollection
func InitializeMovieCollection() {
	if dbs.DB == nil {
		log.Fatal("Database not initialized")
	}
	movieCollection = dbs.DB.Collection("movies")
}

// Hàm này trả về collection của Movie để controller có thể sử dụng lại
func GetMovieCollection() *mongo.Collection {
	return movieCollection
}

// Khởi tạo validator
var validatemovie = validator.New()

// Validate method for Movie struct
func (movie *Movie) Validate() error {
	validate := validator.New()

	// Validate struct fields
	if err := validate.Struct(movie); err != nil {
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
				case "url":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be a valid URL")
				case "len":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be exactly "+fieldErr.Param()+" characters long")
				case "numeric":
					errorMessages = append(errorMessages, fieldErr.Field()+" must be numeric")
				case "dive":
					errorMessages = append(errorMessages, fieldErr.Field()+" contains invalid elements")
				default:
					errorMessages = append(errorMessages, fieldErr.Field()+" is invalid")
				}
			}
			/// Trả về một lỗi tổng hợp từ các thông báo lỗi chi tiết
			return errors.New("Validation failed: " + joinErrorsMovie(errorMessages))
		}
		return err
	}
	return nil
}

// Hàm joinErrors để nối các thông báo lỗi thành một chuỗi
func joinErrorsMovie(errors []string) string {
	return strings.Join(errors, ", ")
}
