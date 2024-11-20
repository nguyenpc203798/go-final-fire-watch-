package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Định nghĩa struct New
type New struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`                // ID của bài viết tin tức
	Name            string             `bson:"name" json:"name"`                       // Tên tiêu đề bài viết
	Status          int                `bson:"status" json:"status"`                   // Trạng thái bài viết (0: nháp, 1: đã xuất bản)
	Image           string             `bson:"image" json:"image"`                     // Ảnh chính
	MoreImage       []string           `bson:"moreimage,omitempty" json:"moreimage"`   // Ảnh phụ
	Description     string             `bson:"description" json:"description"`         // Mô tả ngắn
	MoreDescription string             `bson:"moredescription" json:"moredescription"` // Nội dung chi tiết
	CreatedAt       time.Time          `bson:"created_at,omitempty" json:"created_at"` // Ngày tạo slide
	UpdatedAt       time.Time          `bson:"updated_at,omitempty" json:"updated_at"` // Ngày cập nhật slide
	Category        string             `bson:"category" json:"category"`               // Danh mục bài viết
	Slug            string             `bson:"slug" json:"slug"`                       // URL thân thiện
	Views           int                `bson:"views,omitempty" json:"views"`           // Lượt xem bài viết
}
