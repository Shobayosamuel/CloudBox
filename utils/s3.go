package utils

import (
    "log"
    "os"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

var s3Client *s3.S3

func GetS3Client() *s3.S3 {
    if s3Client != nil {
        return s3Client
    }

    // Initialize AWS session with credentials
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(GetEnv("AWS_REGION")),
        Credentials: credentials.NewStaticCredentials(
            GetEnv("AWS_ACCESS_KEY_ID"),
            GetEnv("AWS_SECRET_ACCESS_KEY"),
            "", // session token
        ),
    })
    if err != nil {
        log.Printf("Failed to create AWS session: %v", err)
        return nil
    }

    s3Client = s3.New(sess)
    return s3Client
}

func GetEnv(key string, defaultValue ...string) string {
    value := os.Getenv(key)
    if value == "" && len(defaultValue) > 0 {
        return defaultValue[0]
    }
    return value
}