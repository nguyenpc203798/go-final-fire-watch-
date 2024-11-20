//models/role.go
package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Định nghĩa struct Role
type Role struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"` // ID của role
	Name   string             `bson:"name" json:"name"`        // Tên của vai trò (admin, customer, v.v.)
	Status int                `bson:"status"`
}
