// Package adapter provides the GORM implementation of the JobRepository.
package adapter

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"selfier/internal/job/core"
)

// GormJob is the GORM model for the Job entity.
// It maps the core.Job struct to a database table, keeping persistence details
// separate from the core domain model.
type GormJob struct {
	gorm.Model
	ID      string          `gorm:"primaryKey;type:varchar(36)"`
	Type    core.JobType    `gorm:"type:varchar(50);not null"`
	Status  core.JobStatus  `gorm:"type:varchar(50);not null;index"`
	Payload json.RawMessage `gorm:"type:jsonb"`
	Result  json.RawMessage `gorm:"type:jsonb"`
	Error   *string         `gorm:"type:text"`
}

func (m *GormJob) toCoreJob() *core.Job {
	if m == nil {
		return nil
	}
	return &core.Job{
		ID:        m.ID,
		Type:      m.Type,
		Status:    m.Status,
		Payload:   m.Payload,
		Result:    m.Result,
		Error:     m.Error,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toCoreJobs(gormJobs []GormJob) []*core.Job {
	coreJobs := make([]*core.Job, len(gormJobs))
	for i, m := range gormJobs {
		job := m
		coreJobs[i] = job.toCoreJob()
	}
	return coreJobs
}

func fromCoreJob(m *core.Job) *GormJob {
	if m == nil {
		return nil
	}
	return &GormJob{ //nolint:exhaustruct
		ID:      m.ID,
		Type:    m.Type,
		Status:  m.Status,
		Payload: m.Payload,
		Result:  m.Result,
		Error:   m.Error,
	}
}

// gormJobRepository is the GORM implementation of the core.JobRepository interface.
type gormJobRepository struct {
	db *gorm.DB
}

// NewGormJobRepository creates a new GORM-backed job repository.
// It also runs AutoMigrate to ensure the jobs table is created or updated.
func NewGormJobRepository(database *gorm.DB) (core.JobRepository, error) {
	// Auto-migrate the schema to create/update the 'gorm_jobs' table
	//nolint:exhaustruct
	if err := database.AutoMigrate(&GormJob{}); err != nil {
		return nil, err
	}
	return &gormJobRepository{db: database}, nil
}

// CreateJob saves a new job to the database.
func (r *gormJobRepository) CreateJob(ctx context.Context, job *core.Job) (*core.Job, error) {
	gormJob := fromCoreJob(job)
	result := r.db.WithContext(ctx).Create(gormJob)
	if result.Error != nil {
		return nil, result.Error
	}
	return gormJob.toCoreJob(), nil
}

// GetAllJobs retrieves all jobs from the database.
func (r *gormJobRepository) GetAllJobs(ctx context.Context) ([]*core.Job, error) {
	var gormJobs []GormJob
	result := r.db.WithContext(ctx).Find(&gormJobs)
	if result.Error != nil {
		return nil, result.Error
	}

	return toCoreJobs(gormJobs), nil
}

// GetJobByID retrieves a single job by its ID.
func (r *gormJobRepository) GetJobByID(ctx context.Context, id string) (*core.Job, error) {
	var gormJob GormJob
	result := r.db.WithContext(ctx).First(&gormJob, "id = ?", id)
	if result.Error != nil {
		// This will correctly return gorm.ErrRecordNotFound if no record is found.
		return nil, result.Error
	}
	return gormJob.toCoreJob(), nil
}

// UpdateJobStatus updates an existing job's status, result, and error fields.
func (r *gormJobRepository) UpdateJobStatus(ctx context.Context, id string, status core.JobStatus, result json.RawMessage, errMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
		"result": result,
		"error":  errMsg,
	}

	//nolint:exhaustruct
	dbResult := r.db.WithContext(ctx).Model(&GormJob{}).
		Where("id = ?", id).
		Updates(updates)

	return dbResult.Error
}

// DeleteJobByID deletes a job from the database by its ID.
func (r *gormJobRepository) DeleteJobByID(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&GormJob{}, "id = ?", id) //nolint:exhaustruct
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
