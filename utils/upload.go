package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadToS3(filePath string) (string, error) {

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not find dir, %v", err)
	}

	file, err := os.Open(dir + filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file, %v", err)
	}
	defer file.Close()

	objectKey := filepath.Base(filePath)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"),
	})
	if err != nil {
		return "", fmt.Errorf("could not start session, %v", err)
	}

	svc := s3.New(sess)


	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("manim-bot"),
		Key: aws.String(objectKey),
		Body: file,
		ContentType: aws.String("video/mp4"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to s3: %v", err)
	}

	file.Close()

	err = os.RemoveAll(dir + "/static")
	if err != nil {
		return "", fmt.Errorf("file uploaded to s3 but failed to delete local file: %w", err)
	}
	
	return fmt.Sprintf("https://manim-bot.s3.ap-south-1.amazonaws.com/%s",objectKey),nil
}