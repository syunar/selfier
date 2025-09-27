// Package job contains the ports or interefaces for the job module
package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusFinished = "finished"
	StatusFailed   = "failed"
)

const (
	JobTypeDeselfie = "deselfie"
)

// Domain

type Job struct {
	ID          string
	Type        string
	ModelConfig map[string]interface{}
	ImageKey    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type JobResult struct {
	ID                string
	JobID             string
	ImagePresignedURL *string
	Stataus           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

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

// DTOs

type CreateJobRawBody struct {
	Type        string        `form:"type" doc:"Job type" enum:"deselfie" example:"deselfie" required:"true"`
	ModelConfig string        `form:"model_config" doc:"Job model configuration" example:"{\"prompt\": \"a photo of a person\"}" default:"{}" required:"true"`
	Image       huma.FormFile `form:"file" required:"true"`
}

type CreateJobInput struct {
	RawBody huma.MultipartFormFiles[CreateJobRawBody]
}

type JobBody struct {
	ID          string                 `json:"id" example:"123"`
	Type        string                 `json:"type" example:"deselfie"`
	ModelConfig map[string]interface{} `json:"model_config" example:"{\"prompt\": \"a photo of a person\"}"`
	ImageKey    string                 `json:"image_key" example:"jobs/123/image.jpg"`
	Status      string                 `json:"status" example:"pending"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type JobOutput struct {
	Body JobBody `json:"body"`
}

type JobsOutput struct {
	Body []JobBody `json:"body"`
}

// Model

type JobModel struct {
	gorm.Model
	ID          string `gorm:"type:uuid;primaryKey"`
	Type        string
	ModelConfig datatypes.JSON `gorm:"type:jsonb"`
	ImageKey    string
	Status      string `gorm:"type:varchar(20);not null;default:'pending'"`
}

type JobResultModel struct {
	gorm.Model
	ID        string `gorm:"type:uuild;primaryKey"`
	JobID     string `gorm:"type:uuid"`
	ImageKey  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Ports

type JobHTTPHandler interface {
	CreateJob(ctx context.Context, input *CreateJobInput) (*JobOutput, error)
	GetJobs(ctx context.Context, input *struct{}) (*JobsOutput, error)
	GetJobByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*JobOutput, error)
	DeleteJobByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*struct{}, error)
	GetJobResultsByID(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*JobResultsOutput, error)
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
	CreateJob(ctx context.Context, id string, jobType string, modelConfig map[string]interface{}, imageKey string, status string) (*JobModel, error)
	GetJobs(ctx context.Context) ([]JobModel, error)
	GetJobByID(ctx context.Context, id string) (*JobModel, error)
	DeleteJobByID(ctx context.Context, id string) error
	UpdateJobStatus(ctx context.Context, id string, status string) (*JobModel, error)
}

type JobResultRepository interface {
	CreateResult(ctx context.Context, id string, jobID string, imageKey string) (*JobResultModel, error)
	GetResultsByJobID(ctx context.Context, jobID string) ([]JobResultModel, error)
}

type JobObjectStorage interface {
	UploadImage()
	GetPresignedURL()
}

type JobEventPublisher interface {
	PublishCreatedJobEvent()
}
