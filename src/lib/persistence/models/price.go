package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Price ...
type Price struct {
	ID                primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	TheaterID         primitive.ObjectID `json:"cinema_id,omitempty" bson:"theaterId"`
	Weight            uint               `json:"-" bson:"weight"`
	Label             string             `json:"label" bson:"label"`
	Full              float32            `json:"full,omitempty" bson:"full"`
	Half              float32            `json:"half,omitempty" bson:"half"`
	IncludingPreviews bool               `json:"includingPreviews" bson:"includingPreviews"`
	IncludingHolidays bool               `json:"includingHolidays" bson:"includingHolidays"`
	ExceptPreviews    bool               `json:"exceptPreviews" bson:"exceptPreviews"`
	ExceptHolidays    bool               `json:"exceptHolidays" bson:"exceptHolidays"`
	Weekdays          []time.Weekday     `json:"weekdays,omitempty" bson:"weekdays"`
	Attributes        []string           `json:"attributes" bson:"attributes"`
	CreatedAt         *time.Time         `json:"-" bson:"createdAt"`
	Theater           *Theater           `json:"cinema,omitempty" bson:"theater,omitempty"`
}
