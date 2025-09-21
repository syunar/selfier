// Package core
package core

import (
	"context"
	"encoding/json"
	"fmt"

	deselfieCore "selfier/internal/deselfie/core"

	"github.com/google/uuid"
)

type jobServiceImpl struct {
	repo             JobRepository
	blobRepo         BlobRepository
	deselfiePipeline deselfieCore.DeselfiePipeline
}

func NewJobService(repo JobRepository, blobRepo BlobRepository, deselfiePipeline deselfieCore.DeselfiePipeline) JobService {
	return &jobServiceImpl{
		repo:             repo,
		blobRepo:         blobRepo,
		deselfiePipeline: deselfiePipeline,
	}
}

// CreateJob handles the business logic for creating a new job.
// It generates a unique ID, sets the initial status to 'pending',
// and then persists it using the repository.
func (s *jobServiceImpl) CreateJob(ctx context.Context, jobType JobType, payload json.RawMessage) (*Job, error) {
	// 1. Create a new Job struct with initial values.
	newJob := &Job{ //nolint:exhaustruct
		ID:      uuid.New().String(), // Generate a unique ID for the job.
		Type:    jobType,
		Status:  StatusPending, // New jobs always start as pending.
		Payload: payload,
	}

	// 2. Persist the new job using the repository.
	createdJob, err := s.repo.CreateJob(ctx, newJob)
	if err != nil {
		return nil, fmt.Errorf("failed to create job in repository: %w", err)
	}

	// 3. (Optional) Enqueue the job for a background worker.
	// In a real-world scenario, after successfully creating the job record,
	// you would publish an event or push the job ID to a message queue (like RabbitMQ, SQS, or Redis)
	// for a separate worker process to pick up and execute.
	// Example: s.jobQueueClient.Enqueue(ctx, createdJob.ID)

	return createdJob, nil
}

// GetAllJobs retrieves all jobs by delegating the call directly to the repository.
func (s *jobServiceImpl) GetAllJobs(ctx context.Context) ([]*Job, error) {
	jobs, err := s.repo.GetAllJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all jobs: %w", err)
	}
	return jobs, nil
}

// GetJobByID retrieves a single job by its ID by delegating the call to the repository.
func (s *jobServiceImpl) GetJobByID(ctx context.Context, id string) (*Job, error) {
	job, err := s.repo.GetJobByID(ctx, id)
	if err != nil {
		// You might want to handle 'not found' errors specifically here
		// and return a custom application-level error.
		return nil, fmt.Errorf("failed to get job by id %s: %w", id, err)
	}
	return job, nil
}

// DeleteJobByID deletes a job by its ID by delegating the call to the repository.
func (s *jobServiceImpl) DeleteJobByID(ctx context.Context, id string) error {
	err := s.repo.DeleteJobByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete job by id %s: %w", id, err)
	}
	return nil
}
