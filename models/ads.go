package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Định nghĩa struct Ads
type Ads struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`                  // ID của quảng cáo
	Title       string             `bson:"title" json:"title"`                       // Tiêu đề của quảng cáo
	Image       string             `bson:"image" json:"image"`                       // Đường dẫn đến hình ảnh quảng cáo
	Link        string             `bson:"link,omitempty" json:"link"`               // Liên kết mà quảng cáo trỏ tới
	Position    string             `bson:"position,omitempty" json:"position"`       // Vị trí hiển thị của quảng cáo (banner, sidebar, etc.)
	Status      int                `bson:"status" json:"status"`                     // Trạng thái (0: không hiển thị, 1: hiển thị)
	Description string             `bson:"description,omitempty" json:"description"` // Mô tả cho quảng cáo (nếu có)
	StartDate   time.Time          `bson:"start_date,omitempty" json:"start_date"`   // Ngày bắt đầu hiển thị quảng cáo
	EndDate     time.Time          `bson:"end_date,omitempty" json:"end_date"`       // Ngày kết thúc hiển thị quảng cáo
	CreatedAt   time.Time          `bson:"created_at,omitempty" json:"created_at"`   // Ngày tạo quảng cáo
	UpdatedAt   time.Time          `bson:"updated_at,omitempty" json:"updated_at"`   // Ngày cập nhật quảng cáo
}


      