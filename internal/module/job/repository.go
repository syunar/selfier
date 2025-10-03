package job

import (
	"context"
	"encoding/json"
	"log/slog"
	"selfier/pkg/middleware"

	"gorm.io/gorm"
)

type jobRepositoryGorm struct {
	db *gorm.DB
}

func NewJobRepositoryGorm(db *gorm.DB) (JobRepository, error) {
	//nolint:exhaustruct
	err := db.AutoMigrate(&JobModel{})
	if err != nil {
		return nil, err
	}
	return &jobRepositoryGorm{db: db}, nil
}

func (r *jobRepositoryGorm) CreateJob(ctx context.Context, id string, jobType string, modelConfig map[string]interface{}, status string) (*JobModel, error) {

	log := middleware.GetLogger(ctx)

	configBytes, err := json.Marshal(modelConfig)
	if err != nil {
		log.Error("failed to marshal model config", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	jobModel := JobModel{ //nolint:exhaustruct
		ID:          id,
		Type:        jobType,
		Status:      status,
		ModelConfig: configBytes,
	}
	result := r.db.WithContext(ctx).Create(&jobModel)

	if result.Error != nil {
		log.Error("failed to create job", slog.String("job_id", id), slog.String("error", result.Error.Error()))
		if result.RowsAffected == 0 && result.Error == gorm.ErrDuplicatedKey {
			return nil, ErrConflict
		}
		return nil, ErrInternal
	}

	return &jobModel, nil
}

func (r *jobRepositoryGorm) GetJobs(ctx context.Context) ([]*JobModel, error) {
	log := middleware.GetLogger(ctx)

	var jobs []*JobModel

	result := r.db.WithContext(ctx).Find(&jobs)
	if result.Error != nil {
		log.Error("failed to get jobs", slog.String("error", result.Error.Error()))
		return nil, ErrInternal
	}

	return jobs, nil
}

func (r *jobRepositoryGorm) GetJobByID(ctx context.Context, id string) (*JobModel, error) {
	log := middleware.GetLogger(ctx)
	var job JobModel

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&job)
	if result.Error != nil {
		log.Error("failed to get job by ID", slog.String("job_id", id), slog.String("error", result.Error.Error()))
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &job, nil
}

func (r *jobRepositoryGorm) DeleteJobByID(ctx context.Context, id string) error {
	log := middleware.GetLogger(ctx)

	//nolint:exhaustruct
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&JobModel{})

	if result.Error != nil {
		log.Error("failed to delete job", slog.String("job_id", id), slog.String("error", result.Error.Error()))
		return ErrInternal
	}

	return nil
}

func (r *jobRepositoryGorm) UpdateJobStatus(ctx context.Context, id string, status string) (*JobModel, error) {
	log := middleware.GetLogger(ctx)

	//nolint:exhaustruct
	result := r.db.WithContext(ctx).Model(&JobModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": status})

	if result.Error != nil {
		log.Error("failed to update job status", slog.String("job_id", id), slog.String("status", status), slog.String("error", result.Error.Error()))
		return nil, ErrInternal
	}

	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return r.GetJobByID(ctx, id)
}

type JobImageRepositoryGorm struct {
	db *gorm.DB
}

func NewJobImageRepositoryGorm(db *gorm.DB) (JobImageRepository, error) {
	//nolint:exhaustruct
	err := db.AutoMigrate(&JobImageModel{})
	if err != nil {
		return nil, err
	}
	return &JobImageRepositoryGorm{db: db}, nil
}

func (r *JobImageRepositoryGorm) CreateImage(ctx context.Context, id string, jobID string, imageKey string, imageType string) (*JobImageModel, error) {
	log := middleware.GetLogger(ctx)

	jobImage := &JobImageModel{ //nolint:exhaustruct
		ID:       id,
		JobID:    jobID,
		ImageKey: imageKey,
		Type:     imageType,
	}

	result := r.db.WithContext(ctx).Create(jobImage)
	if result.Error != nil {
		log.Error("failed to create job result", slog.String("result_id", id), slog.String("job_id", jobID), slog.String("error", result.Error.Error()))
		if result.RowsAffected == 0 && result.Error == gorm.ErrDuplicatedKey {
			return nil, ErrConflict
		}
		return nil, ErrInternal
	}

	return jobImage, nil
}

func (r *JobImageRepositoryGorm) GetImages(ctx context.Context, jobID string) ([]*JobImageModel, error) {
	log := middleware.GetLogger(ctx)
	var jobImageModels []*JobImageModel

	result := r.db.WithContext(ctx).Where("job_id = ?", jobID).Find(&jobImageModels)
	if result.Error != nil {
		log.Error("failed to get job results by job ID", slog.String("job_id", jobID), slog.String("error", result.Error.Error()))
		return nil, ErrInternal
	}

	return jobImageModels, nil
}

func (r *JobImageRepositoryGorm) GetImage(ctx context.Context, jobID string, imageID string) (*JobImageModel, error) {
	log := middleware.GetLogger(ctx)
	var jobImageModel JobImageModel

	result := r.db.WithContext(ctx).Where("job_id = ?", jobID).Where("id = ?", imageID).First(&jobImageModel)
	if result.Error != nil {
		log.Error("failed to get job result by ID", slog.String("job_id", jobID), slog.String("image_id", imageID), slog.String("error", result.Error.Error()))
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &jobImageModel, nil
}
