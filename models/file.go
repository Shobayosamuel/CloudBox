package models

import (
    "time"
    "gorm.io/gorm"
)

type File struct {
    gorm.Model
    UserID      uint      `json:"user_id"`
    FileName    string    `json:"file_name"`
    FileSize    int64     `json:"file_size"`
    ContentType string    `json:"content_type"`
    CloudPath   string    `json:"cloud_path"`
    UploadDate  time.Time `json:"upload_date"`
    User        User      `gorm:"foreignKey:UserID"`
}