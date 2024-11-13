package controllers

import (
	"CloudBox/models"
	"CloudBox/utils"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
    MinPasswordLength = 8
    MaxPasswordLength = 72
    MaxLoginAttempts = 5
    LockoutDuration = 15 * time.Minute
)

type AuthInput struct {
    Username string `json:"username" binding:"required,min=3,max=30"`
    Password string `json:"password" binding:"required,min=8,max=72"`
    Email    string `json:"email" binding:"required,email"`
}

func validatePassword(password string) error {
    if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
        return fmt.Errorf("password must be between %d and %d characters", MinPasswordLength, MaxPasswordLength)
    }

    hasUpper := false
    hasLower := false
    hasNumber := false
    hasSpecial := false

    for _, char := range password {
        switch {
        case 'a' <= char && char <= 'z':
            hasLower = true
        case 'A' <= char && char <= 'Z':
            hasUpper = true
        case '0' <= char && char <= '9':
            hasNumber = true
        case strings.ContainsRune("!@#$%^&*", char):
            hasSpecial = true
        }
    }

    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain at least one uppercase letter, lowercase letter, number, and special character")
    }

    return nil
}

func CreateUser(c *gin.Context) {
    db := utils.ConnectDB()
    var input AuthInput

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate password
    if err := validatePassword(input.Password); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check if username exists
    var existingUser models.User
    if result := db.Where("username = ?", input.Username).First(&existingUser); result.Error == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
        return
    }

    // Check if email exists
    if result := db.Where("email = ?", input.Email).First(&existingUser); result.Error == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
        return
    }

    user := models.User{
        Username: input.Username,
        Password: string(hashedPassword),
        Email:    input.Email,
    }

    if result := db.Create(&user); result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}


func Login(c *gin.Context) {
    db := utils.ConnectDB()
    var input AuthInput

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if result := db.Where("username = ?", input.Username).First(&user); result.Error != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    // Check account lockout
    if user.LockedUntil.After(time.Now()) {
        c.JSON(http.StatusTooManyRequests, gin.H{
            "error": fmt.Sprintf("account is locked. Try again after %v", user.LockedUntil),
        })
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        user.LoginAttempts++

        // Lock account if too many attempts
        if user.LoginAttempts >= MaxLoginAttempts {
            user.LockedUntil = time.Now().Add(LockoutDuration)
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "account locked due to too many failed attempts"})
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentialtems"})
        }

        db.Save(&user)
        return
    }

    // Reset login attempts on successful login
    user.LoginAttempts = 0
    user.LastLogin = time.Now()
    db.Save(&user)

    // Generate tokens
    tokens, err := utils.GenerateTokens(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
        return
    }
    response := gin.H{
        "tokens": tokens,
        "user": gin.H{
            "id": user.ID,
            "username": user.Username,
            "lastLogin": user.LastLogin,
            "login-attempts": user.LoginAttempts,
        },
    }

    c.JSON(http.StatusOK, response)
}

func RefreshToken(c *gin.Context) {
    refreshToken := c.GetHeader("Refresh-Token")
    if refreshToken == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
        return
    }

    token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(os.Getenv("SECRET")), nil
    })

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
        return
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if claims["type"] != "refresh" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
            return
        }

        userID := uint(claims["user_id"].(float64))
        tokens, err := utils.GenerateTokens(userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
            return
        }

        c.JSON(http.StatusOK, tokens)
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
    }
}

func GetUserProfile(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    db := utils.ConnectDB()
    var user models.User
    if result := db.First(&user, userID); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "username": user.Username,
        "email":    user.Email,
        "lastLogin": user.LastLogin,
    })
}