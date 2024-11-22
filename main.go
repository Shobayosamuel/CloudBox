package main

import (
    "fmt"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "CloudBox/controllers"
    "CloudBox/middlewares"
    "CloudBox/utils"
    "time"
)

func main() {
    // Load env variables
    utils.LoadEnv()

    r := gin.Default()

    // Detailed CORS configuration
    r.Use(cors.New(cors.Config{
        AllowAllOrigins:  true, // Temporarily allow all origins
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
        AllowHeaders:     []string{
            "Origin",
            "Content-Length",
            "Content-Type",
            "Authorization",
            "Refresh-Token",
            "Accept",
        },
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    // Logging middleware for debugging
    r.Use(func(c *gin.Context) {
        fmt.Printf("Request Origin: %s\n", c.GetHeader("Origin"))
        fmt.Printf("Request Method: %s\n", c.Request.Method)
        c.Next()
    })

    // Public routes
    auth := r.Group("/auth")
    {
        auth.OPTIONS("/register", func(c *gin.Context) {
            c.Header("Access-Control-Allow-Origin", "*")
            c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
            c.Status(200)
        })
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