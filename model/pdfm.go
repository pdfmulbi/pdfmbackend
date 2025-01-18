package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PdfmUsers struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email" json:"email"`
	Password      string             `bson:"password" json:"password"`
	IsSupport     bool               `bson:"isSupport" json:"isSupport"`
	LastMergeTime time.Time          `bson:"lastMergeTime" json:"lastMergeTime"`
	MergeCount    int                `bson:"mergeCount" json:"mergeCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
