package job

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"selfier/pkg/middleware"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type jobObjectStorageS3 struct {
	s3Client *s3.Client
	bucket   string
}

func NewJobObjectStorageS3(s3 *s3.Client, bucket string) JobObjectStorage {
	return &jobObjectStorageS3{s3, bucket}
}

func (o *jobObjectStorageS3) UploadImage(ctx context.Context, fileName string, file io.Reader) (string, error) {

	log := middleware.GetLogger(ctx)

	fileKey := fmt.Sprintf("%s_%d", fileName, time.Now().Unix())

	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, file)
	if err != nil {
		log.Error("failed to read file", slog.String("error", err.Error()))
		return "", ErrReadFile
	}

	_, err = o.s3Client.PutObject(ctx, &s3.PutObjectInput{ //nolint:exhaustruct
		Bucket:      aws.String(o.bucket),
		Key:         aws.String(fileKey),
		Body:        bytes.NewReader(buffer.Bytes()),
		ContentType: aws.String(http.DetectContentType(buffer.Bytes())),
	})
	if err != nil {
		log.Error("failed to upload file to S3", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	return fileKey, nil
}

func (o *jobObjectStorageS3) GetPresignedURL(ctx context.Context, objectKey string) (string, error) {

	log := middleware.GetLogger(ctx)

	presignClient := s3.NewPresignClient(o.s3Client)

	// TODO: make it configurable
	expiration := time.Minute * 60

	presignedURL, err := presignClient.PresignGetObject(ctx,
		&s3.GetObjectInput{ //nolint:exhaustruct
			Bucket: aws.String(o.bucket),
			Key:    aws.String(objectKey),
		},
		s3.WithPresignExpires(expiration))

	if err != nil {
		log.Error("failed to create presigned URL", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	return presignedURL.URL, nil
}

func (o *jobObjectStorageS3) DeleteImage(ctx context.Context, key string) error {

	log := middleware.GetLogger(ctx)

	_, err := o.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{ //nolint:exhaustruct
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Error("failed to delete object", slog.String("error", err.Error()))
		return ErrInternal
	}

	return nil
}
