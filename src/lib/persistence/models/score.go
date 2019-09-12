package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Score ...
type Score struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	MovieID primitive.ObjectID `json:"movieId" bson:"movieId,omitempty"`
	Imdb    struct {
		ID    string  `json:"id" bson:"id,omitempty"`
		Score float32 `json:"score" bson:"score,omitempty"`
	} `json:"imdb" bson:"imdb,omitempty"`
	Rotten struct {
		Path  string `json:"path" bson:"path,omitempty"`
		Class string `json:"class" bson:"class,omitempty"`
		Score int    `json:"score" bson:"score,omitempty"`
	} `json:"rotten" bson:"rotten,omitempty"`
	KeepSynced bool       `json:"keepSynced,omitempty" bson:"keepSynced"`
	CreatedAt  *time.Time `json:"created_at,omitempty" bson:"createdAt"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty" bson:"updatedAt"`
}
