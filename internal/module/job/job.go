// Package job contains the ports or interefaces for the job module
package job

import (
	"context"
	"io"
)

type JobHTTPHandler interface {
	CreateJob(ctx context.Context, input *CreateJobInput) (*JobOutput, error)
	GetJobs(ctx context.Context, input *struct{}) (*JobsOutput, error)
	GetJobByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*JobOutput, error)
	DeleteJobByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*struct{}, error)
	GetJobImagesByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*JobImagesOutput, error)
	GetPresignedURL(ctx context.Context, input *GetPresignedURLInput) (*GetPresignedURLOutput, error)
}

type JobService interface {
	CreateJob(ctx context.Context, jobType string, modelConfig map[string]interface{}, imageReader io.Reader, imageFilename string) (*Job, error)
	GetJobs(ctx context.Context) ([]*Job, error)
	GetJobByID(ctx context.Context, jobID string) (*Job, error)
	DeleteJobByID(ctx context.Context, jobID string) error
	GetJobImagesByID(ctx context.Context, jobID string) ([]*JobImage, error)
	UpdateJobStatus(ctx context.Context, jobID string, status string) (*Job, error)
	UploadImage(ctx context.Context, imageReader io.Reader, imageFilename string) (string, error)
	GetPresignedURL(ctx context.Context, jobID string, imageID string) (string, error)
}

type JobRepository interface {
	CreateJob(ctx context.Context, id string, jobType string, modelConfig map[string]interface{}, status string) (*JobModel, error)
	GetJobs(ctx context.Context) ([]*JobModel, error)
	GetJobByID(ctx context.Context, id string) (*JobModel, error)
	DeleteJobByID(ctx context.Context, id string) error
	UpdateJobStatus(ctx context.Context, id string, status string) (*JobModel, error)
}

type JobImageRepository interface {
	CreateImage(ctx context.Context, imageID string, jobID string, imageKey string, imageType string) (*JobImageModel, error)
	GetImages(ctx context.Context, jobID string) ([]*JobImageModel, error)
	GetImage(ctx context.Context, jobID string, imageID string) (*JobImageModel, error)
}

type JobObjectStorage interface {
	UploadImage(ctx context.Context, fileName string, file io.Reader) (string, error)
	GetPresignedURL(ctx context.Context, objectKey string) (string, error)
	DeleteImage(ctx context.Context, key string) error
}

type JobEventPublisher interface {
	PublishCreatedJobEvent()
}
