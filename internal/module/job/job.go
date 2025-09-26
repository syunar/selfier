// Package job contains the ports or interefaces for the job module
package job

import (
	"context"
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
	CreateJob()
	GetJobs()
	GetJobByID()
	DeleteJobByID()
	GetJobResultsByID()
	UpdateJobStatus()
	UploadImage()
	GetPresignedURL()
}

type JobRepository interface {
	CreateJob()
	GetJobs()
	GetJobByID()
	DeleteJobByID()
	GetJobResultsByID()
	UpdateJobStatus()
}

type JobResultRepository interface {
	CreateResult()
	GetResultsByJobID()
}

type JobObjectStorage interface {
	UploadImage()
	GetPresignedURL()
}

type JobEventPublisher interface {
	PublishCreatedJobEvent()
}
