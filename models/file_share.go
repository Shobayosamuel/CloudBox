package models

import (
    "time"
    "gorm.io/gorm"
)

type FileShare struct {
    gorm.Model
    FileID      uint      `json:"file_id"`
    ShareToken  string    `json:"share_token" gorm:"unique"`
    CreatedBy   uint      `json:"created_by"`
    ExpiresAt   time.Time `json:"expires_at"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    AccessCount int       `json:"access_count" gorm:"default:0"`
    File        File      `gorm:"foreignKey:FileID"`
}