package job

import (
	"context"
	"io"
	"log/slog"
	"selfier/pkg/middleware"

	"github.com/google/uuid"
)

type jobServiceImpl struct {
	jobRepository      JobRepository
	jobImageRepository JobImageRepository
	jobObjectStorage   JobObjectStorage
	jobEventPublisher  JobEventPublisher
}

func NewJobService(
	jobRepository JobRepository,
	jobImageRepository JobImageRepository,
	jobObjectStorage JobObjectStorage,
	jobEventPublisher JobEventPublisher,
) JobService {
	return &jobServiceImpl{
		jobRepository:      jobRepository,
		jobImageRepository: jobImageRepository,
		jobObjectStorage:   jobObjectStorage,
		jobEventPublisher:  jobEventPublisher,
	}
}

func (s *jobServiceImpl) CreateJob(
	ctx context.Context,
	jobType string,
	modelConfig map[string]interface{},
	imageReader io.Reader,
	imageFilename string,
) (*Job, error) {

	log := middleware.GetLogger(ctx)

	jobID := uuid.New().String()
	imageKey, err := s.UploadImage(ctx, imageReader, imageFilename)
	if err != nil {
		log.Error("failed to upload image", slog.String("error", err.Error()))
		if err == ErrReadFile {
			return nil, ErrReadFile
		}
		return nil, ErrInternal
	}

	jobModel, err := s.jobRepository.CreateJob(ctx, jobID, jobType, modelConfig, StatusPending)
	if err != nil {
		log.Error("failed to create job", slog.String("error", err.Error()))
		if err == ErrConflict {
			return nil, ErrConflict
		}
		return nil, ErrInternal
	}

	imageID := uuid.New().String()
	_, err = s.jobImageRepository.CreateImage(ctx, imageID, jobModel.ID, imageKey, ImageTypeInput)

	if err != nil {
		log.Error("failed to create image", slog.String("error", err.Error()))
		if err == ErrConflict {
			return nil, ErrConflict
		}
		return nil, ErrInternal
	}

	_, err = s.jobEventPublisher.JobCreated(ctx, &JobCreatedEventData{
		JobID:        jobModel.ID,
		InputImageID: imageID,
		ModelConfig:  modelConfig,
	})

	if err != nil {
		log.Error("failed to publish job created event", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

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
		log.Error(
			"failed to get job by ID",
			slog.String("job_id", jobID),
			slog.String("error", err.Error()),
		)
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
		log.Error(
			"failed to delete job by ID",
			slog.String("job_id", jobID),
			slog.String("error", err.Error()),
		)
		return ErrInternal
	}

	return nil
}

func (s *jobServiceImpl) GetJobImagesByID(ctx context.Context, jobID string) ([]*JobImage, error) {

	log := middleware.GetLogger(ctx)

	JobImageModels, err := s.jobImageRepository.GetImages(ctx, jobID)
	if err != nil {
		log.Error(
			"failed to get job results by job ID",
			slog.String("job_id", jobID),
			slog.String("error", err.Error()),
		)
		return nil, ErrInternal
	}

	JobImages := make([]*JobImage, len(JobImageModels))
	for i, JobImageModel := range JobImageModels {
		JobImage, _ := GetJobImageFromModel(JobImageModel)
		JobImages[i] = JobImage
	}

	return JobImages, nil
}

func (s *jobServiceImpl) GetJobImageByID(
	ctx context.Context,
	jobID string,
	imageID string,
) (*JobImage, error) {

	log := middleware.GetLogger(ctx)

	JobImageModel, err := s.jobImageRepository.GetImage(ctx, jobID, imageID)
	if err != nil {
		log.Error(
			"failed to get job result by ID",
			slog.String("job_id", jobID),
			slog.String("image_id", imageID),
			slog.String("error", err.Error()),
		)
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	JobImage, err := GetJobImageFromModel(JobImageModel)
	if err != nil {
		log.Error("failed to get job image from model", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	return JobImage, nil
}

func (s *jobServiceImpl) UpdateJobStatus(
	ctx context.Context,
	jobID string,
	status string,
) (*Job, error) {
	// s.jobRepository.UpdateJobStatus()
	log := middleware.GetLogger(ctx)

	jobModel, err := s.jobRepository.UpdateJobStatus(ctx, jobID, status)
	if err != nil {
		log.Error(
			"failed to update job status",
			slog.String("job_id", jobID),
			slog.String("status", status),
			slog.String("error", err.Error()),
		)
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

func (s *jobServiceImpl) UploadImage(
	ctx context.Context,
	imageReader io.Reader,
	imageFilename string,
) (string, error) {
	log := middleware.GetLogger(ctx)

	fileKey, err := s.jobObjectStorage.UploadImage(ctx, imageFilename, imageReader)
	if err != nil {
		log.Error("failed to upload image", slog.String("error", err.Error()))
		if err == ErrReadFile {
			return "", ErrReadFile
		}
		return "", ErrInternal
	}
	return fileKey, nil
}

func (s *jobServiceImpl) GetPresignedURL(
	ctx context.Context,
	jobID string,
	imageID string,
) (string, error) {

	log := middleware.GetLogger(ctx)

	jobImageModel, err := s.jobImageRepository.GetImage(ctx, jobID, imageID)
	if err != nil {
		log.Error(
			"failed to get job image by ID",
			slog.String("job_id", jobID),
			slog.String("image_id", imageID),
			slog.String("error", err.Error()),
		)
		if err == ErrNotFound {
			return "", ErrNotFound
		}
		return "", ErrInternal
	}

	presignedURL, err := s.jobObjectStorage.GetPresignedURL(ctx, jobImageModel.ImageKey)
	if err != nil {
		log.Error("failed to create presigned url", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	return presignedURL, nil
}
