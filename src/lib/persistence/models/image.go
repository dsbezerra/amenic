package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImageType string

const (
	// ImageTypeAny is used when the image is not categorized
	ImageTypeAny ImageType = "any"
	// ImageTypeBackdrop is used to indicate backdrop images
	ImageTypeBackdrop ImageType = "backdrop"
	// ImageTypePoster is used to indicate poster images
	ImageTypePoster ImageType = "poster"
)

// Image ...
type Image struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	MovieID   primitive.ObjectID `json:"movieId,omitempty" bson:"movieId,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Path      string             `json:"path,omitempty" bson:"path,omitempty"`
	Type      string             `json:"type,omitempty" bson:"type,omitempty"`
	Main      bool               `json:"main,omitempty" bson:"main"`
	URL       string             `json:"url,omitempty" bson:"url,omitempty"`
	SecureURL string             `json:"secure_url,omitempty" bson:"secure_url,omitempty"`
	Width     int                `json:"width,omitempty" bson:"width,omitempty"`
	Height    int                `json:"height,omitempty" bson:"height,omitempty"`
	Checksum  string             `json:"checksum,omitempty" bson:"checksum,omitempty"` // SHA1 Checksum
	CreatedAt *time.Time         `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt *time.Time         `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	Palette   *Palette           `json:"palette,omitempty" bson:"palette,omitempty"`
	Movie     *Movie             `json:"movie,omitempty" bson:"movie,omitempty"`
}

// Palette ...
type Palette struct {
	VibrantColor      string `json:"vibrant_color,omitempty" bson:"vibrantColor,omitempty"`
	LightVibrantColor string `json:"light_vibrant_color,omitempty" bson:"lightVibrantColor,omitempty"`
	DarkVibrantColor  string `json:"dark_vibrant_color,omitempty" bson:"darkVibrantColor,omitempty"`
	MutedColor        string `json:"muted_color,omitempty" bson:"mutedColor,omitempty"`
	LightMutedColor   string `json:"light_muted_color,omitempty" bson:"lightMutedColor,omitempty"`
	DarkMutedColor    string `json:"dark_muted_color,omitempty" bson:"darkMutedColor,omitempty"`
}

// CheckForValidImageType ...
func CheckForValidImageType(it *ImageType) bool {
	if it == nil {
		return false
	}

	if *it == "" {
		*it = ImageTypeAny
	}

	switch *it {
	case ImageTypeAny, ImageTypeBackdrop, ImageTypePoster:
		return true
	default:
		return false
	}
}
