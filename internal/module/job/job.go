// Package job contains the ports or interefaces for the job module
package job

import "context"

type JobHTTPHandler interface {
	CreateJob(ctx context.Context, input *struct{ any })
	GetJobs(ctx context.Context, input *struct{ any })
	GetJobByID(ctx context.Context, input *struct{ any })
	DeleteJobByID(ctx context.Context, input *struct{ any })
	GetJobResultsByID(ctx context.Context, input *struct{ any })
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
