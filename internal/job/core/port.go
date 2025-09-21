// Package core
package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

type JobType string

const (
	JobTypeDeselfie JobType = "deselfie"
)

// JobStatus represents the lifecycle of a job.
type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type Job struct {
	ID        string          `json:"id"`
	Type      JobType         `json:"type"`
	Status    JobStatus       `json:"status"`
	Payload   json.RawMessage `json:"payload"`          // Holds the specific input data (e.g., DeselfiePayload)
	Result    json.RawMessage `json:"result,omitempty"` // Holds the specific output data (e.g., DeselfieResult)
	Error     *string         `json:"error,omitempty"`  // To store error messages on failure
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// JobHTTPHandler defines the HTTP layer for managing jobs.
type JobHTTPHandler interface {
	CreateJob(c *gin.Context)
	GetAllJobs(c *gin.Context)
	GetJobByID(c *gin.Context)
	DeleteJobByID(c *gin.Context)
}

// BlobRepository defines a generic interface for storing and retrieving binary data.
type BlobRepository interface {
	Save(ctx context.Context, key string, data []byte) (string, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// JobRepository defines the persistence layer for Job metadata.
type JobRepository interface {
	CreateJob(ctx context.Context, job *Job) (*Job, error)
	GetAllJobs(ctx context.Context) ([]*Job, error)
	GetJobByID(ctx context.Context, id string) (*Job, error)
	UpdateJobStatus(ctx context.Context, id string, status JobStatus, result json.RawMessage, errMsg *string) error
	DeleteJobByID(ctx context.Context, id string) error
}

// JobService defines the business logic for managing jobs.
type JobService interface {
	CreateJob(ctx context.Context, jobType JobType, payload json.RawMessage) (*Job, error)
	GetAllJobs(ctx context.Context) ([]*Job, error)
	GetJobByID(ctx context.Context, id string) (*Job, error)
	UpdateJobStatus(ctx context.Context, id string, status JobStatus, result json.RawMessage, errMsg *string) error
	DeleteJobByID(ctx context.Context, id string) error
}
