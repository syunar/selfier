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

const (
	BucketName string = "selfier"
)

type storages3 struct {
	s3Client *s3.Client
}

func NewStorageS3(s3Client *s3.Client) Storage {
	return &storages3{s3Client}
}

func (s *storages3) Upload(ctx context.Context, file io.Reader, fileName string) (string, error) {

	log := middleware.GetLogger(ctx)

	fileKey := fmt.Sprintf("%s_%d", fileName, time.Now().Unix())
	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, file)
	if err != nil {
		log.Error("failed to read file", slog.String("error", err.Error()))
		return "", ErrUnreadableFile
	}
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(BucketName),
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

func (s *storages3) GetPresignedURL(ctx context.Context, key string) (string, error) {

	log := middleware.GetLogger(ctx)

	presignClient := s3.NewPresignClient(s.s3Client)

	expiration := time.Minute * 60

	presignedURL, err := presignClient.PresignGetObject(ctx,
		&s3.GetObjectInput{ //nolint:exhaustruct
			Bucket: aws.String(BucketName),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(expiration))

	if err != nil {
		log.Error("failed to create presigned URL", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	return presignedURL.URL, nil
}

func (s *storages3) Delete(ctx context.Context, key string) error {

	log := middleware.GetLogger(ctx)
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{ //nolint:exhaustruct
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Error("failed to delete object", slog.String("error", err.Error()))
		return ErrInternal
	}

	return nil
}
