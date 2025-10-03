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

	// ErrReadFile is returned when a file cannot be read.
	ErrReadFile = errors.New("failed to read file")
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

const (
	ImageTypeInput  = "input"
	ImageTypeOutput = "output"
)

// ---------------
// Domain Entities
// ---------------

type Job struct {
	ID          string                 `json:"id" example:"123"`
	Type        string                 `json:"type" example:"deselfie"`
	ModelConfig map[string]interface{} `json:"model_config" example:"{\"prompt\": \"a photo of a person\"}"`
	Status      string                 `json:"status" example:"pending"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type JobImage struct {
	ID        string    `json:"id" example:"abc"`
	JobID     string    `json:"job_id" example:"123"`
	ImageKey  string    `json:"image_key" example:"jobs/123/image.jpg"`
	Type      string    `json:"type" example:"input"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		Status:      jobModel.Status,
		CreatedAt:   jobModel.CreatedAt,
		UpdatedAt:   jobModel.UpdatedAt,
	}, nil
}

func GetJobImageFromModel(jobImageModel *JobImageModel) (*JobImage, error) {
	return &JobImage{ //nolint:exhaustruct
		ID:        jobImageModel.ID,
		JobID:     jobImageModel.JobID,
		ImageKey:  jobImageModel.ImageKey,
		Type:      jobImageModel.Type,
		CreatedAt: jobImageModel.CreatedAt,
		UpdatedAt: jobImageModel.UpdatedAt,
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

type GetPresignedURLOutputBody struct {
	URL string `json:"url"`
}

// ---- Input and Output ----

type CreateJobInput struct {
	RawBody huma.MultipartFormFiles[CreateJobRawBody]
}

type GetPresignedURLInput struct {
	ID    string `path:"image_id" required:"true" doc:"Image ID" example:"f6b57be8-aa17-4e46-8cca-396cb7f977e9"`
	JobID string `path:"job_id" required:"true" doc:"Job ID" example:"052ef2f5-28dd-44e6-8345-e8fc11235d1c"`
}

type GetPresignedURLOutput struct {
	Body GetPresignedURLOutputBody `json:"body"`
}

type JobOutput struct {
	Body *Job `json:"body"`
}

type JobsOutput struct {
	Body []*Job `json:"body"`
}

type JobImagesOutput struct {
	Body []*JobImage `json:"body"`
}

// ---------------
// Repository Models
// ---------------

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type JobModel struct {
	Model
	ID          string         `gorm:"type:uuid;primaryKey"`
	Type        string         `gorm:"type:string"`
	ModelConfig datatypes.JSON `gorm:"type:jsonb"`
	Status      string         `gorm:"type:varchar(20);not null;default:'pending'"`
}

type JobImageModel struct {
	Model
	ID       string `gorm:"type:uuid;primaryKey"`
	JobID    string `gorm:"type:uuid;not null;index"`
	ImageKey string `gorm:"type:string;not null"`
	Type     string `gorm:"type:string;not null"`
}
