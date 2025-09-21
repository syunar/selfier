// Package core
package core

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// JobStatus represents the lifecycle of a job.
type JobStatus string

const (
	JobTypeDeselfie JobType = "deselfie"
)

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type JobType string

type Job struct {
	ID      string          `json:"id"`
	Type    JobType         `json:"type"`
	Status  JobStatus       `json:"status"`
	Payload json.RawMessage `json:"payload"`          // Holds the specific input data (e.g., DeselfiePayload)
	Result  json.RawMessage `json:"result,omitempty"` // Holds the specific output data (e.g., DeselfieResult)
	Error   *string         `json:"error,omitempty"`  // To store error messages on failure
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
	UpdateJob(ctx context.Context, job *Job) error
	DeleteJobByID(ctx context.Context, id string) error
}

// JobService defines the business logic for managing jobs.
type JobService interface {
	CreateJob(ctx context.Context, jobType JobType, payload json.RawMessage) (*Job, error)
	GetAllJobs(ctx context.Context) ([]*Job, error)
	GetJobByID(ctx context.Context, id string) (*Job, error)
	DeleteJobByID(ctx context.Context, id string) error
}
