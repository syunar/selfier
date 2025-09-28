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
	GetJobResultsByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*JobResultsOutput, error)
}

type JobService interface {
	CreateJob(ctx context.Context, jobType string, modelConfig map[string]interface{}, imageReader io.Reader, imageFilename string) (*Job, error)
	GetJobs(ctx context.Context) ([]*Job, error)
	GetJobByID(ctx context.Context, jobID string) (*Job, error)
	DeleteJobByID(ctx context.Context, jobID string) error
	GetJobResultsByID(ctx context.Context, jobID string) ([]*JobResult, error)
	// UpdateJobStatus(ctx context.Context, jobID string, status string) (*Job, error)
	// UploadImage(ctx context.Context, imageReader io.Reader, imageFilename string) (string, error)
	// GetPresignedURL(ctx context.Context, imageKey string) (string, error)
}

type JobRepository interface {
	CreateJob(ctx context.Context, id string, jobType string, modelConfig map[string]interface{}, imageKey string, status string) (*JobModel, error)
	GetJobs(ctx context.Context) ([]JobModel, error)
	GetJobByID(ctx context.Context, id string) (*JobModel, error)
	DeleteJobByID(ctx context.Context, id string) error
	UpdateJobStatus(ctx context.Context, id string, status string) (*JobModel, error)
}

type JobResultRepository interface {
	CreateResult(ctx context.Context, id string, jobID string, imageKey string) (*JobResultModel, error)
	GetResultsByJobID(ctx context.Context, jobID string) ([]JobResultModel, error)
}

type JobObjectStorage interface {
	UploadImage()
	GetPresignedURL()
}

type JobEventPublisher interface {
	PublishCreatedJobEvent()
}
