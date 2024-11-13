package utils

import (
    "log"
    "os"
    "time"
    "net"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func ConnectDB() *gorm.DB {
    // Add retry logic for DNS resolution
    maxRetries := 3
    var db *gorm.DB
    var err error

    dsn := os.Getenv("DB_URL")
    if dsn == "" {
        log.Fatal("DB_URL environment variable is not set")
    }

    // Configure GORM with optimized settings
    config := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        PrepareStmt: true,
    }

    // Retry loop for connection
    for i := 0; i < maxRetries; i++ {
        // Try to resolve the host first
        host := "ep-crimson-truth-a5ttv8vs.us-east-2.aws.neon.tech"
        _, err := net.LookupHost(host)
        if err != nil {
            log.Printf("Attempt %d: DNS resolution failed: %v", i+1, err)
            time.Sleep(2 * time.Second) // Wait before retrying
            continue
        }

        db, err = gorm.Open(postgres.Open(dsn), config)
        if err == nil {
            break
        }
        log.Printf("Attempt %d: Failed to connect to database: %v", i+1, err)
        time.Sleep(2 * time.Second) // Wait before retrying
    }

    if err != nil {
        log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
    }

    // Configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatal("Failed to get database instance:", err)
    }

    // Configure connection pool settings
    sqlDB.SetMaxIdleConns(5)
    sqlDB.SetMaxOpenConns(20)
    sqlDB.SetConnMaxLifetime(time.Hour)

    // Verify connection
    err = sqlDB.Ping()
    if err != nil {
        log.Fatal("Failed to ping database:", err)
    }

    log.Println("Successfully connected to database!")
    return db
}