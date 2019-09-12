package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification ...
type Notification struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Type       string             `json:"type" bson:"type,omitempty"`
	Title      string             `json:"title" bson:"title,omitempty"` // Content of notification to be displayed to user
	Text       string             `json:"text" bson:"text,omitempty"`
	HTMLText   string             `json:"htmlText" bson:"htmlText,omitempty"`
	Single     bool               `json:"single" bson:"single,omitempty"` // Whether this is a notification of a single item or not, if true, ItemID holds the resource identifier.
	ItemID     string             `json:"itemId" bson:"itemId,omitempty"`
	NowPlaying *string            `json:"nowPlaying" bson:"nowPlaying,omitempty"` // NowPlaying field is defined only if notification type is "premiere"
	Data       *NotificationData  `json:"data,omitempty" bson:"data,omitempty"`   // Data contains extra information
	CreatedAt  time.Time          `json:"createdAt,omitempty" bson:"createdAt"`
}

// NotificationData is used to store data payload of notification
type NotificationData struct {
	Movies   []Movie   `json:"movies,omitempty" bson:"movies,omitempty"`
	Theaters []Theater `json:"cinemas,omitempty" bson:"cinemas,omitempty"`
}
