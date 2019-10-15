package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Admin struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id"` // Document id.
	Username  string             `json:"username" bson:"username"` // Username
	Password  string             `json:"password" bson:"password"` // Password
	CreatedAt *time.Time         `json:"created_at,omitempty" bson:"created_at,omitempty"`
}
