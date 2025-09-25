package job

import (
	"context"
)

type jobHTTPHandlerImpl struct {
	JobService JobService
}

func NewJobHTTPHandler(jobService JobService) JobHTTPHandler {
	return &jobHTTPHandlerImpl{
		JobService: jobService,
	}
}

func (h *jobHTTPHandlerImpl) CreateJob(ctx context.Context, input *struct{ any }) {
	h.JobService.CreateJob()
}

func (h *jobHTTPHandlerImpl) GetJobs(ctx context.Context, input *struct{ any }) {
	h.JobService.GetJobs()
}

func (h *jobHTTPHandlerImpl) GetJobByID(ctx context.Context, input *struct{ any }) {
	h.JobService.GetJobByID()
}

func (h *jobHTTPHandlerImpl) DeleteJobByID(ctx context.Context, input *struct{ any }) {
	h.JobService.DeleteJobByID()
}

func (h *jobHTTPHandlerImpl) GetJobResultsByID(ctx context.Context, input *struct{ any }) {
	h.JobService.GetJobResultsByID()
}
