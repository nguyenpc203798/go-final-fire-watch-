package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Định nghĩa struct Slide
type Slide struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`                  // ID của slide
	Title       string             `bson:"title" json:"title"`                       // Tiêu đề của slide
	Image       string             `bson:"image" json:"image"`                       // Đường dẫn đến hình ảnh của slide
	Link        string             `bson:"link,omitempty" json:"link"`               // Liên kết mà slide trỏ đến (ví dụ khi người dùng click vào)
	Status      int                `bson:"status" json:"status"`                     // Trạng thái (0: ẩn, 1: hiển thị)
	Position    int                `bson:"position" json:"position"`                 // Vị trí của slide trong slideshow
	Description string             `bson:"description,omitempty" json:"description"` // Mô tả cho slide (nếu cần)
	CreatedAt   time.Time          `bson:"created_at,omitempty" json:"created_at"`   // Ngày tạo slide
	UpdatedAt   time.Time          `bson:"updated_at,omitempty" json:"updated_at"`   // Ngày cập nhật slide
}
