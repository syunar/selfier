package job

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"selfier/pkg/middleware"

	"github.com/google/uuid"
)

type jobServiceImpl struct {
	jobRepository       JobRepository
	jobResultRepository JobResultRepository
	// jobObjectStorage    JobObjectStorage
	// jobEventPublisher   JobEventPublisher
}

func NewJobService(jobRepository JobRepository, jobResultRepository JobResultRepository) JobService {
	return &jobServiceImpl{
		jobRepository:       jobRepository,
		jobResultRepository: jobResultRepository,
		// jobEventPublisher:   jobEventPublisher,
		// jobObjectStorage:    jobObjectStorage,
	}
}

func (s *jobServiceImpl) CreateJob(ctx context.Context, jobType string, modelConfig map[string]interface{}, imageReader io.Reader, imageFilename string) (*Job, error) {

	log := middleware.GetLogger(ctx)

	id := uuid.New().String()

	// TODO
	// s.jobObjectStorage.UploadImage()
	imageKey := "jobs/" + id + "/" + imageFilename

	jobModel, err := s.jobRepository.CreateJob(ctx, id, jobType, modelConfig, imageKey, StatusPending)
	if err != nil {
		log.Error("failed to create job", slog.String("error", err.Error()))
		if err == ErrConflict {
			return nil, ErrConflict
		}
		return nil, ErrInternal
	}

	// TODO
	// s.jobEventPublisher.PublishCreatedJobEvent()

	job, err := GetJobFromModel(jobModel)
	if err != nil {
		log.Error("failed to get job from model", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	return job, nil
}

func (s *jobServiceImpl) GetJobs(ctx context.Context) ([]*Job, error) {
	log := middleware.GetLogger(ctx)

	jobsModel, err := s.jobRepository.GetJobs(ctx)
	if err != nil {
		log.Error("failed to get jobs", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	jobs := make([]*Job, len(jobsModel))
	for i, jobModel := range jobsModel {
		job, err := GetJobFromModel(jobModel)
		if err != nil {
			log.Error("failed to get job from model", slog.String("error", err.Error()))
			return nil, ErrInternal
		}
		jobs[i] = job
	}

	return jobs, nil
}

func (s *jobServiceImpl) GetJobByID(ctx context.Context, jobID string) (*Job, error) {

	log := middleware.GetLogger(ctx)

	jobModel, err := s.jobRepository.GetJobByID(ctx, jobID)

	if err != nil {
		log.Error("failed to get job by ID", slog.String("job_id", jobID), slog.String("error", err.Error()))
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	job, err := GetJobFromModel(jobModel)
	if err != nil {
		log.Error("failed to get job from model", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	return job, nil
}

func (s *jobServiceImpl) DeleteJobByID(ctx context.Context, jobID string) error {
	// s.jobRepository.DeleteJobByID()
	log := middleware.GetLogger(ctx)

	err := s.jobRepository.DeleteJobByID(ctx, jobID)
	if err != nil {
		log.Error("failed to delete job by ID", slog.String("job_id", jobID), slog.String("error", err.Error()))
		return ErrInternal
	}

	return nil
}

func (s *jobServiceImpl) GetJobResultsByID(ctx context.Context, jobID string) ([]*JobResult, error) {

	log := middleware.GetLogger(ctx)

	jobResultModels, err := s.jobResultRepository.GetResultsByJobID(ctx, jobID)
	if err != nil {
		log.Error("failed to get job results by job ID", slog.String("job_id", jobID), slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	jobResults := make([]*JobResult, len(jobResultModels))
	for i, jobResultModel := range jobResultModels {
		jobResult, _ := GetJobResultFromModel(jobResultModel)
		jobResults[i] = jobResult
	}

	// TODO
	// get presigned url

	// attach presigned url

	// TODO
	for i, jobResult := range jobResults {
		jobResults[i].ImagePresignedURL, _ = s.GetPresignedURL(ctx, jobResult.ImageKey)
		// if err != nil {
		// 	log.Error("failed to get presigned url", slog.String("error", err.Error()))
		// 	return nil, err
		// }
	}

	return jobResults, nil
}

func (s *jobServiceImpl) UpdateJobStatus(ctx context.Context, jobID string, status string) (*Job, error) {
	// s.jobRepository.UpdateJobStatus()
	log := middleware.GetLogger(ctx)

	jobModel, err := s.jobRepository.UpdateJobStatus(ctx, jobID, status)
	if err != nil {
		log.Error("failed to update job status", slog.String("job_id", jobID), slog.String("status", status), slog.String("error", err.Error()))
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	job, err := GetJobFromModel(jobModel)
	if err != nil {
		log.Error("failed to get job from model", slog.String("error", err.Error()))
		return nil, ErrInternal
	}
	return job, nil
}

// func (s *jobServiceImpl) UploadImage(ctx context.Context, imageReader io.Reader, imageFilename string) (string, error) {
// 	// s.jobObjectStorage.UploadImage()
// 	return "", nil
// }

func (s *jobServiceImpl) GetPresignedURL(ctx context.Context, imageKey string) (string, error) {
	// s.jobObjectStorage.GetPresignedURL()
	return fmt.Sprintf("https://s3.amazonaws.com/%s", imageKey), nil
}
