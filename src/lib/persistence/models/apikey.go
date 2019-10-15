package models

import (
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// UserTypeAdmin ...
	UserTypeAdmin = "admin"

	// UserTypeClient ...
	UserTypeClient = "client"
)

// APIKey represents an API key document.
type APIKey struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id"`             // Document id.
	Key       string             `json:"key,omitempty" bson:"key"`             // Actual key.
	Name      string             `json:"name,omitempty" bson:"name"`           // Name used to identify key.
	UserType  string             `json:"user_type,omitempty" bson:"user_type"` // Which user type is this key such as admin, client or whatever else we need (which we will not).
	Platform  string             `json:"platform,omitempty" bson:"platform"`   // Which platform is using this key
	Owner     string             `json:"owner,omitempty" bson:"owner"`         // Whoever created this key, probably the username.
	Timestamp *time.Time         `json:"iat,omitempty" bson:"iat"`             // The time when the key was created.
}

// UserTypeLevel ...
func (m APIKey) UserTypeLevel() int {
	return UserTypeLevel(m.UserType)
}

// UserTypeLevel ...
func UserTypeLevel(t string) int {
	result := 0

	switch t {
	case UserTypeAdmin:
		result = math.MaxInt16 // any big number

	case UserTypeClient:
		result = 1
	}

	return result
}
