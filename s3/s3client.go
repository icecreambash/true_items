package s3

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

type S3Client struct {
	svc    *s3.S3
	bucket string
}

// NewS3Client создает новый экземпляр S3Client
func NewS3Client() (*S3Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("S3_BUCKET_NAME")
	endpoint := os.Getenv("AWS_ENDPOINT")

	sess, _ := session.NewSession()

	svc := s3.New(sess, aws.NewConfig().WithEndpoint(endpoint).WithRegion(region))

	return &S3Client{
		svc: svc,
		//svc:    s3.New(sess),
		bucket: bucket,
	}, nil
}

// UploadFile загружает файл в S3
func (c *S3Client) UploadFile(filePath string, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = c.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

// DownloadFile загружает файл из S3
func (c *S3Client) DownloadFile(key string, destPath string) error {
	result, err := c.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}
	defer result.Body.Close()

	outFile, err := os.Create(destPath)

	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, result.Body)
	return err
}

// GetFilePath возвращает путь к файлу в S3
func (c *S3Client) GetFilePath(tenant string, folder string, key string) string {
	url := os.Getenv("AWS_URL")

	return fmt.Sprintf("%s/tenants/tenant%s/app/public/%s/%s", url, tenant, folder, key)
}
