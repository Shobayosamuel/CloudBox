package main

import (
    "CloudBox/controllers"
    "CloudBox/middlewares"
    "CloudBox/utils"
    "fmt"
    "os"
    "time"
    "net/http"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    // Load env variables
    utils.LoadEnv()

    r := gin.Default()

    // CORS configuration
    config := cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Refresh-Token", "Accept"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:          12 * time.Hour,
    }

    // Add CORS middleware
    r.Use(cors.New(config))

    // Add OPTIONS handler for all routes
    r.OPTIONS("/*path", func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, Refresh-Token, Accept")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Status(http.StatusOK)
    })

    // Logging middleware for debugging
    r.Use(func(c *gin.Context) {
        c.Next()
    })

    // Public routes
    auth := r.Group("/auth")
    {
        auth.POST("/register", controllers.CreateUser)
        auth.POST("/login", controllers.Login)
        auth.POST("/refresh", controllers.RefreshToken)
    }

    // Protected routes
    protected := r.Group("/api")
    protected.Use(middlewares.CheckAuth())
    {
        protected.GET("/profile", controllers.GetUserProfile)
        protected.POST("/files/upload", controllers.UploadFile)
        protected.GET("/files/list", controllers.ListFiles)
        protected.GET("/files/download/:id", controllers.DownloadFile)
    }

    // Use port from environment variable
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    fmt.Printf("Server starting on port %s\n", port)
    r.Run(":" + port)
}