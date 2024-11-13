package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username       string    `json:"username" gorm:"unique"`
    Password       string    `json:"password"`
    Email         string    `json:"email" gorm:"unique"`
    LoginAttempts int       `json:"login_attempts" gorm:"default:0"`
    LockedUntil   time.Time `json:"locked_until"`
    LastLogin     time.Time `json:"last_login"`
}