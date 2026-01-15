package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MergeHistory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"65b..."`
	UserID     primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	InputFiles []string           `bson:"input_files" json:"input_files" example:"file1.pdf,file2.pdf"`
	OutputFile string             `bson:"output_file" json:"output_file" example:"merged_result.pdf"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type CompressHistory struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"65b..."`
	UserID         primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FileName       string             `bson:"file_name" json:"file_name" example:"laporan_besar.pdf"`
	OriginalSize   int64              `bson:"original_size" json:"original_size" example:"5000000"`
	CompressedSize int64              `bson:"compressed_size" json:"compressed_size" example:"1500000"`
	Status         string             `bson:"status" json:"status" example:"success"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

type ConvertHistory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"65b..."`
	UserID       primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FileName     string             `bson:"file_name" json:"file_name" example:"document.docx"`
	SourceFormat string             `bson:"source_format" json:"source_format" example:"docx"`
	TargetFormat string             `bson:"target_format" json:"target_format" example:"pdf"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type SummaryHistory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"65b..."`
	UserID      primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FileName    string             `bson:"file_name" json:"file_name" example:"jurnal.pdf"`
	SummaryText string             `bson:"summary_text" json:"summary_text" example:"Ringkasan dokumen ini adalah..."`
	Language    string             `bson:"language" json:"language" example:"id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// untuk Swagger
type MergeInput struct {
	InputFiles []string `json:"input_files" example:"file1.pdf,file2.pdf"`
	OutputFile string   `json:"output_file" example:"merged_result.pdf"`
}

type CompressInput struct {
	FileName       string `json:"file_name" example:"laporan_besar.pdf"`
	OriginalSize   int64  `json:"original_size" example:"5000000"`
	CompressedSize int64  `json:"compressed_size" example:"1500000"`
	Status         string `json:"status" example:"success"`
}

type ConvertInput struct {
	FileName     string `json:"file_name" example:"document.docx"`
	SourceFormat string `json:"source_format" example:"docx"`
	TargetFormat string `json:"target_format" example:"pdf"`
}

type SummaryInput struct {
	FileName    string `json:"file_name" example:"jurnal.pdf"`
	SummaryText string `json:"summary_text" example:"Ringkasan..."`
	Language    string `json:"language" example:"id"`
}

type DeleteHistoryInput struct {
	ID   string `json:"id" example:"65b123..."`
	Type string `json:"type" example:"merge"` // merge, compress, convert, summary
}

// ==========================================
// 3. RESPONSE STRUCTS (Untuk Unified History)
// ==========================================

type HistoryItem struct {
	ID          string      `json:"id"`
	Type        string      `json:"type" example:"merge"`
	Description string      `json:"description" example:"Merged 2 PDF files"`
	FileName    string      `json:"file_name"`
	Details     interface{} `json:"details,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type UnifiedHistoryResponse struct {
	Status  int           `json:"status" example:"200"`
	Message string        `json:"message" example:"History retrieved successfully"`
	History []HistoryItem `json:"history"`
}

type HistoryActionResponse struct {
	Message string             `json:"message" example:"Log berhasil disimpan"`
	ID      primitive.ObjectID `json:"id,omitempty" example:"65b4..."`
}