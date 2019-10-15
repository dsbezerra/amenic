package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Theater ...
type Theater struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	CityID     primitive.ObjectID `json:"cityId,omitempty" bson:"cityId"`
	Hidden     bool               `json:"hidden,omitempty" bson:"hidden"` // Whether it should be retrieved or not.
	InternalID string             `json:"internalId,omitempty" bson:"internalId"`
	Name       string             `json:"name,omitempty" bson:"name"`
	ShortName  string             `json:"shortName,omitempty" bson:"shortName"`
	Images     *TheaterImages     `json:"images,omitempty" bson:"images"`
	CreatedAt  *time.Time         `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt  *time.Time         `json:"updatedAt,omitempty" bson:"updatedAt"`
	City       *City              `json:"city,omitempty" bson:"city,omitempty"`
	Prices     []Price            `json:"prices,omitempty" bson:"prices,omitempty"`
	Sessions   []Session          `json:"sessions,omitempty" bson:"sessions,omitempty"`
}

// TheaterImages ...
type TheaterImages struct {
	BackdropURL string `json:"backdrop,omitempty" bson:"backdrop"`
	IconURL     string `json:"icon,omitempty" bson:"icon"`
	LogoURL     string `json:"logo,omitempty" bson:"logo"`
}
