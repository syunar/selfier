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

func (r *jobRepositoryGorm) CreateJob(ctx context.Context, id string, jobType string, modelConfig map[string]interface{}, imageKey string, status string) (*JobModel, error) {

	log := middleware.GetLogger(ctx)

	configBytes, err := json.Marshal(modelConfig)
	if err != nil {
		log.Error("failed to marshal model config", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	jobModel := JobModel{ //nolint:exhaustruct
		ID:          id,
		Type:        jobType,
		ImageKey:    imageKey,
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

func (r *jobRepositoryGorm) GetJobs(ctx context.Context) ([]JobModel, error) {
	log := middleware.GetLogger(ctx)

	var jobs []JobModel

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

type jobResultRepositoryGorm struct {
	db *gorm.DB
}

func NewJobResultRepositoryGorm(db *gorm.DB) (JobResultRepository, error) {
	//nolint:exhaustruct
	err := db.AutoMigrate(&JobResultModel{})
	if err != nil {
		return nil, err
	}
	return &jobResultRepositoryGorm{db: db}, nil
}

func (r *jobResultRepositoryGorm) CreateResult(ctx context.Context, id string, jobID string, imageKey string) (*JobResultModel, error) {
	log := middleware.GetLogger(ctx)

	jobResult := JobResultModel{ //nolint:exhaustruct
		ID:       id,
		JobID:    jobID,
		ImageKey: imageKey,
	}

	result := r.db.WithContext(ctx).Create(&jobResult)
	if result.Error != nil {
		log.Error("failed to create job result", slog.String("result_id", id), slog.String("job_id", jobID), slog.String("error", result.Error.Error()))
		if result.RowsAffected == 0 && result.Error == gorm.ErrDuplicatedKey {
			return nil, ErrConflict
		}
		return nil, ErrInternal
	}

	return &jobResult, nil
}

func (r *jobResultRepositoryGorm) GetResultsByJobID(ctx context.Context, jobID string) ([]JobResultModel, error) {
	log := middleware.GetLogger(ctx)
	var jobResultModels []JobResultModel

	result := r.db.WithContext(ctx).Where("job_id = ?", jobID).Find(&jobResultModels)
	if result.Error != nil {
		log.Error("failed to get job results by job ID", slog.String("job_id", jobID), slog.String("error", result.Error.Error()))
		return nil, ErrInternal
	}

	return jobResultModels, nil
}
