package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ---------------
// Errors
// ---------------

var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("job record not found")

	// ErrConflict is returned when an operation violates a unique constraint
	// (e.g., trying to create a job with an ID that already exists).
	ErrConflict = errors.New("job record conflict, probably non-unique ID")

	// ErrInternal is a general error for unexpected failures (e.g., marshaling, DB connection loss).
	ErrInternal = errors.New("internal data access error")
)

// ---------------
// Constants
// ---------------

const (
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusFinished = "finished"
	StatusFailed   = "failed"
)

const (
	JobTypeDeselfie = "deselfie"
)

// ---------------
// Domain Entities
// ---------------

type Job struct {
	ID          string                 `json:"id" example:"123"`
	Type        string                 `json:"type" example:"deselfie"`
	ModelConfig map[string]interface{} `json:"model_config" example:"{\"prompt\": \"a photo of a person\"}"`
	ImageKey    string                 `json:"image_key" example:"jobs/123/image.jpg"`
	Status      string                 `json:"status" example:"pending"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type JobResult struct {
	ID                string    `json:"id" example:"abc"`
	JobID             string    `json:"job_id" example:"123"`
	ImagePresignedURL string    `json:"image_presigned_url" example:"https://example.com/image.jpg"`
	ImageKey          string    `json:"image_key" example:"jobs/123/image.jpg"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ---------------
// Domain Helper
// ---------------

func GetJobFromModel(jobModel *JobModel) (*Job, error) {
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(jobModel.ModelConfig), &configMap); err != nil {
		return nil, fmt.Errorf("cannot unmarshal model modelConfig: %w", err)
	}
	return &Job{
		ID:          jobModel.ID,
		Type:        jobModel.Type,
		ModelConfig: configMap,
		ImageKey:    jobModel.ImageKey,
		Status:      jobModel.Status,
		CreatedAt:   jobModel.CreatedAt,
		UpdatedAt:   jobModel.UpdatedAt,
	}, nil
}

func GetJobResultFromModel(jobResultModel *JobResultModel) (*JobResult, error) {
	return &JobResult{ //nolint:exhaustruct
		ID:        jobResultModel.ID,
		JobID:     jobResultModel.JobID,
		ImageKey:  jobResultModel.ImageKey,
		CreatedAt: jobResultModel.CreatedAt,
		UpdatedAt: jobResultModel.UpdatedAt,
	}, nil
}

// ---------------
// Handler DTOs
// ---------------

// ---- Body ----

type CreateJobRawBody struct {
	Type        string        `form:"type" doc:"Job type" enum:"deselfie" example:"deselfie" required:"true"`
	ModelConfig string        `form:"model_config" doc:"Job model configuration" example:"{\"prompt\": \"a photo of a person\"}" default:"{}" required:"true"`
	Image       huma.FormFile `form:"image" required:"true"`
}

// ---- Input and Output ----

type CreateJobInput struct {
	RawBody huma.MultipartFormFiles[CreateJobRawBody]
}

type JobOutput struct {
	Body Job `json:"body"`
}

type JobsOutput struct {
	Body []*Job `json:"body"`
}

type JobResultsOutput struct {
	Body []*JobResult `json:"body"`
}

// ---------------
// Repository Models
// ---------------

type JobModel struct {
	gorm.Model
	ID          string         `gorm:"type:uuid;primaryKey"`
	Type        string         `gorm:"type:string"`
	ModelConfig datatypes.JSON `gorm:"type:jsonb"`
	ImageKey    string         `gorm:"type:string"`
	Status      string         `gorm:"type:varchar(20);not null;default:'pending'"`
}

type JobResultModel struct {
	gorm.Model
	ID       string `gorm:"type:uuild;primaryKey"`
	JobID    string `gorm:"type:uuid;not null;index"`
	ImageKey string `gorm:"type:string;not null"`
}
