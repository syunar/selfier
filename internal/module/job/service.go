package job

import (
	"context"
	"io"
)

type jobServiceImpl struct {
	jobRepository       JobRepository
	jobResultRepository JobResultRepository
	jobEventPublisher   JobEventPublisher
	jobObjectStorage    JobObjectStorage
}

func NewJobService(jobRepository JobRepository, jobResultRepository JobResultRepository, jobEventPublisher JobEventPublisher, jobObjectStorage JobObjectStorage) JobService {
	return &jobServiceImpl{
		jobRepository:       jobRepository,
		jobResultRepository: jobResultRepository,
		jobEventPublisher:   jobEventPublisher,
		jobObjectStorage:    jobObjectStorage,
	}
}

func (s *jobServiceImpl) CreateJob(ctx context.Context, jobType string, modelConfig map[string]interface{}, imageReader io.Reader, imageFilename string) (*Job, error) {
	// s.jobRepository.CreateJob()
	// s.jobEventPublisher.PublishCreatedJobEvent()
	return nil, nil
}

func (s *jobServiceImpl) GetJobs(ctx context.Context) ([]*Job, error) {
	// s.jobRepository.GetJobs()
	return nil, nil
}

func (s *jobServiceImpl) GetJobByID(ctx context.Context, jobID string) (*Job, error) {
	// s.jobRepository.GetJobByID()
	return nil, nil
}

func (s *jobServiceImpl) DeleteJobByID(ctx context.Context, jobID string) error {
	// s.jobRepository.DeleteJobByID()
	return nil
}

func (s *jobServiceImpl) GetJobResultsByID(ctx context.Context, jobID string) ([]*JobResult, error) {
	// s.jobResultRepository.GetResultsByJobID()
	return nil, nil
}

func (s *jobServiceImpl) UpdateJobStatus(ctx context.Context, jobID string, status string) (*Job, error) {
	// s.jobRepository.UpdateJobStatus()
	return nil, nil
}

func (s *jobServiceImpl) UploadImage(ctx context.Context, imageReader io.Reader, imageFilename string) (string, error) {
	// s.jobObjectStorage.UploadImage()
	return "", nil
}

func (s *jobServiceImpl) GetPresignedURL(ctx context.Context, imageKey string) (string, error) {
	// s.jobObjectStorage.GetPresignedURL()
	return "", nil
}
