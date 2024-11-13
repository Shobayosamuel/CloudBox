package controllers

import (
    "CloudBox/models"
    "CloudBox/utils"
    "fmt"
    "net/http"
    "path/filepath"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"
)

const (
    MaxFileSize = 100 << 20 // 100 MB
)

type FileUploadResponse struct {
	FileID uint `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int64 `json:"file_size"`
	ContentType string `json:"content_type"`
	UploadDate time.Time `json:"upload_date"`
}

func UploadFile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := c.Request.ParseMultipartForm(MaxFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file was provided"})
		return
	}
	defer file.Close()

	filename := fmt.Sprintf("%s-%s", uuid.New().String(), filepath.Base(header.Filename))
	s3Client := utils.GetS3Client()
    bucket := aws.String(utils.GetEnv("AWS_BUCKET_NAME"))

    // Upload to S3
    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket:        bucket,
        Key:           aws.String(filename),
        Body:          file,
        ContentType:   aws.String(header.Header.Get("Content-Type")),
        ContentLength: aws.Int64(header.Size),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
        return
    }

    // Save file metadata to database
    db := utils.ConnectDB()
    fileRecord := models.File{
        UserID:      userID.(uint),
        FileName:    header.Filename,
        FileSize:    header.Size,
        ContentType: header.Header.Get("Content-Type"),
        CloudPath:   filename,
        UploadDate:  time.Now(),
    }

    if result := db.Create(&fileRecord); result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file metadata"})
        return
    }

    // Return response
    c.JSON(http.StatusOK, FileUploadResponse{
        FileID:      fileRecord.ID,
        FileName:    fileRecord.FileName,
        FileSize:    fileRecord.FileSize,
        ContentType: fileRecord.ContentType,
        UploadDate:  fileRecord.UploadDate,
    })

}

func ListFiles(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user is not authenticated"})
		return
	}
	db := utils.ConnectDB()
	var files []models.File
	if result := db.Where("user_id = ?", userID).Find(&files); result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch files"})
        return
    }

    c.JSON(http.StatusOK, files)
}

func DownloadFile(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    fileID := c.Param("id")

    // Get file metadata from database
    db := utils.ConnectDB()
    var file models.File

    if result := db.Where("id = ? AND user_id = ?", fileID, userID).First(&file); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
        return
    }

    // Generate presigned URL for download
    s3Client := utils.GetS3Client()
    req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(utils.GetEnv("AWS_BUCKET_NAME")),
        Key:    aws.String(file.CloudPath),
    })

    // URL valid for 15 minutes
    url, err := req.Presign(15 * time.Minute)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download url"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "download_url": url,
        "file_name":   file.FileName,
        "expires_in":  "15 minutes",
    })
}
