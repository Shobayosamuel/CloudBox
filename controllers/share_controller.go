package controllers

import (
    "CloudBox/models"
    "CloudBox/utils"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"
)

type CreateShareRequest struct {
    FileID     uint      `json:"file_id" binding:"required"`
    ExpiresIn  int       `json:"expires_in"` // Duration in hours, 0 means no expiration
}

type ShareResponse struct {
    ShareToken  string    `json:"share_token"`
    ShareURL    string    `json:"share_url"`
    ExpiresAt   time.Time `json:"expires_at,omitempty"`
    FileInfo    struct {
        FileName    string    `json:"file_name"`
        FileSize    int64     `json:"file_size"`
        ContentType string    `json:"content_type"`
    } `json:"file_info"`
}

// CreateShareLink generates a new share link for a file
func CreateShareLink(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    var req CreateShareRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db := utils.ConnectDB()

    // Verify file ownership
    var file models.File
    if result := db.Where("id = ? AND user_id = ?", req.FileID, userID).First(&file); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "file not found or access denied"})
        return
    }

    // Generate share token
    shareToken := uuid.New().String()

    // Calculate expiration time
    var expiresAt time.Time
    if req.ExpiresIn > 0 {
        expiresAt = time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
    } else {
        // Set a far future date if no expiration
        expiresAt = time.Now().AddDate(10, 0, 0) // 10 years from now
    }

    // Create share record
    share := models.FileShare{
        FileID:     req.FileID,
        ShareToken: shareToken,
        CreatedBy:  userID.(uint),
        ExpiresAt:  expiresAt,
        IsActive:   true,
    }

    if result := db.Create(&share); result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share link"})
        return
    }

    // Generate share URL
    baseURL := utils.GetEnv("APP_BASE_URL")
    shareURL := fmt.Sprintf("%s/share/%s", baseURL, shareToken)

    response := ShareResponse{
        ShareToken: shareToken,
        ShareURL:   shareURL,
        ExpiresAt:  expiresAt,
        FileInfo: struct {
            FileName    string    `json:"file_name"`
            FileSize    int64     `json:"file_size"`
            ContentType string    `json:"content_type"`
        }{
            FileName:    file.FileName,
            FileSize:    file.FileSize,
            ContentType: file.ContentType,
        },
    }

    c.JSON(http.StatusOK, response)
}

// AccessSharedFile handles access to shared files
func AccessSharedFile(c *gin.Context) {
    shareToken := c.Param("token")

    db := utils.ConnectDB()
    var share models.FileShare

    // Find active share link
    if result := db.Preload("File").Where("share_token = ? AND is_active = ?",
        shareToken, true).First(&share); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "invalid or expired share link"})
        return
    }

    // Check if share has expired
    if time.Now().After(share.ExpiresAt) {
        share.IsActive = false
        db.Save(&share)
        c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
        return
    }

    // Update access count
    share.AccessCount++
    db.Save(&share)

    // Generate temporary download URL
    s3Client := utils.GetS3Client()
    req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(utils.GetEnv("AWS_BUCKET_NAME")),
        Key:    aws.String(share.File.CloudPath),
    })

    // Generate URL valid for 15 minutes
    url, err := req.Presign(15 * time.Minute)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download url"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "file_name":    share.File.FileName,
        "content_type": share.File.ContentType,
        "file_size":    share.File.FileSize,
        "download_url": url,
        "expires_in":   "15 minutes",
    })
}

// ListShares returns all active share links for a user's files
func ListShares(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    db := utils.ConnectDB()
    var shares []models.FileShare

    if result := db.Preload("File").Where("created_by = ? AND is_active = ?",
        userID, true).Find(&shares); result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch shares"})
        return
    }

    c.JSON(http.StatusOK, shares)
}

// RevokeShare deactivates a share link
func RevokeShare(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    shareToken := c.Param("token")

    db := utils.ConnectDB()
    var share models.FileShare

    // Verify ownership and update share status
    result := db.Where("share_token = ? AND created_by = ?",
        shareToken, userID).First(&share)

    if result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
        return
    }

    share.IsActive = false
    db.Save(&share)

    c.JSON(http.StatusOK, gin.H{"message": "share link revoked successfully"})
}