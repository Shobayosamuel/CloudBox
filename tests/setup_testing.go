package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"bytes"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Required environment variables
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}
	requiredVars := []string{
		"AWS_REGION",
		"AWS_BUCKET_NAME",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
	}

	// Check environment variables
	missingVars := []string{}
	for _, varName := range requiredVars {
		fmt.Println(varName)
		if os.Getenv(varName) == "" {
			missingVars = append(missingVars, varName)
		}
	}

	if len(missingVars) > 0 {
		log.Fatalf("Missing environment variables: %s", strings.Join(missingVars, ", "))
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"", // session token (empty if not using temporary credentials)
		),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Create S3 client
	s3Client := s3.New(sess)

	// Generate unique test file name
	testFileName := fmt.Sprintf("test-upload-%s.txt", uuid.New().String())
	bucketName := os.Getenv("AWS_BUCKET_NAME")

	// Create test file content
	testContent := []byte(fmt.Sprintf("S3 Connection Test - %s", time.Now().Format(time.RFC3339)))

	// Upload test file
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testFileName),
		Body:   bytes.NewReader(testContent),
	})
	if err != nil {
		log.Fatalf("Failed to upload test file: %v", err)
	}
	fmt.Printf("✅ Successfully uploaded test file: %s\n", testFileName)

	// Verify file exists
	_, err = s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testFileName),
	})
	if err != nil {
		log.Fatalf("Failed to verify uploaded file: %v", err)
	}
	fmt.Printf("✅ Successfully verified file exists in bucket\n")

	// Optional: Download and verify content (uncomment if needed)
	/*
	getOutput, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testFileName),
	})
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	downloadedContent, err := ioutil.ReadAll(getOutput.Body)
	if err != nil {
		log.Fatalf("Failed to read downloaded content: %v", err)
	}
	if string(downloadedContent) != string(testContent) {
		log.Fatalf("Downloaded content does not match uploaded content")
	}
	fmt.Println("✅ Successfully downloaded and verified file content")
	*/

	// Optional: Delete test file
	_, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testFileName),
	})
	if err != nil {
		log.Fatalf("Failed to delete test file: %v", err)
	}
	fmt.Printf("✅ Deleted test file: %s\n", testFileName)

	fmt.Println("🎉 S3 Connection Test Completed Successfully!")
}