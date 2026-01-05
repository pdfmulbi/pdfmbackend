package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MergeHistory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	InputFiles []string           `bson:"input_files" json:"input_files"` 
	OutputFile string             `bson:"output_file" json:"output_file"` 
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}


type CompressHistory struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FileName       string             `bson:"file_name" json:"file_name"`
	OriginalSize   int64              `bson:"original_size" json:"original_size"`     
	CompressedSize int64              `bson:"compressed_size" json:"compressed_size"` 
	Status         string             `bson:"status" json:"status"`                  
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

type ConvertHistory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FileName     string             `bson:"file_name" json:"file_name"`         
	SourceFormat string             `bson:"source_format" json:"source_format"` 
	TargetFormat string             `bson:"target_format" json:"target_format"` 
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}


type SummaryHistory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FileName    string             `bson:"file_name" json:"file_name"`
	SummaryText string             `bson:"summary_text" json:"summary_text"` 
	Language    string             `bson:"language" json:"language"`         
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}