package utils

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
	"os"
)

var s3Client *s3.S3

func GetS3Client() *s3.S3 {
    if s3Client != nil {
        return s3Client	
    }

    // Initialize AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(GetEnv("AWS_REGION")),
    })
    if err != nil {
        panic("failed to create AWS session")
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