package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// City represents a city.
type City struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	State     State              `json:"state,omitempty" bson:"state,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name"`
	TimeZone  string             `json:"timeZone,omitempty" bson:"timeZone"`
	CreatedAt *time.Time         `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt *time.Time         `json:"updatedAt,omitempty" bson:"updatedAt"`
}
