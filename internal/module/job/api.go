package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"selfier/pkg/middleware"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

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

type jobHTTPHandlerImpl struct {
	JobService JobService
}

func NewJobHTTPHandler(jobService JobService) JobHTTPHandler {
	return &jobHTTPHandlerImpl{
		JobService: jobService,
	}
}

func (h *jobHTTPHandlerImpl) CreateJob(ctx context.Context, input *CreateJobInput) (*JobOutput, error) {
	formData := input.RawBody.Data()

	var modelConfig map[string]interface{}
	if err := json.Unmarshal([]byte(formData.ModelConfig), &modelConfig); err != nil {
		return nil, fmt.Errorf("cannot unmarshal model modelConfig: %w", err)
	}

	jobID := "123"
	fileExt := filepath.Ext(formData.Image.Filename)
	imageKey := fmt.Sprintf("jobs/%s/image%s", jobID, fileExt)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, formData.Image); err != nil {
		return nil, fmt.Errorf("cannot read uploaded file: %w", err)
	}

	resp := &JobOutput{
		Body: JobBody{
			ID:          jobID,
			Type:        formData.Type,
			ModelConfig: modelConfig,
			ImageKey:    imageKey,
			Status:      "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	return resp, nil
}

func (h *jobHTTPHandlerImpl) GetJobs(ctx context.Context, input *struct{}) (*JobsOutput, error) {

	log := middleware.Getlogger(ctx)

	log.Info("getting jobs", slog.String("user_id", middleware.GetUserID(ctx)), slog.String("request_id", middleware.GetRequestID(ctx)))

	resp := &JobsOutput{
		Body: []JobBody{
			{
				ID:          "123",
				Type:        "deselfie",
				ModelConfig: map[string]interface{}{"prompt": "a photo of a person"},
				ImageKey:    "jobs/123/image.jpg",
				Status:      "pending",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "124",
				Type:        "deselfie",
				ModelConfig: map[string]interface{}{"prompt": "a photo of a cat"},
				ImageKey:    "jobs/124/image.jpg",
				Status:      "pending",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}
	return resp, nil
}

func (h *jobHTTPHandlerImpl) GetJobByID(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*JobOutput, error) {

	resp := &JobOutput{
		Body: JobBody{
			ID:          input.ID,
			Type:        "deselfie",
			ModelConfig: map[string]interface{}{"prompt": "a photo of a person"},
			ImageKey:    "jobs/123/image.jpg",
			Status:      "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	return resp, nil
}

func (h *jobHTTPHandlerImpl) DeleteJobByID(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*struct{}, error) {
	h.JobService.DeleteJobByID()

	return &struct{}{}, nil
}

type JobResultsBody struct {
	ID                string `json:"id" example:"abc"`
	JobID             string `json:"job_id" example:"123"`
	ImagePresignedURL string `json:"image_presigned_url" example:"https://example.com/image.jpg"`
}

type JobResultsOutput struct {
	Body []JobResultsBody `json:"body"`
}

func (h *jobHTTPHandlerImpl) GetJobResultsByID(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*JobResultsOutput, error) {
	h.JobService.GetJobResultsByID()
	resp := &JobResultsOutput{
		Body: []JobResultsBody{

			{
				ID:                "abc",
				JobID:             input.ID,
				ImagePresignedURL: "https://example.com/image_abc.jpg",
			},
			{
				ID:                "def",
				JobID:             input.ID,
				ImagePresignedURL: "https://example.com/image_def.jpg",
			},
		},
	}
	return resp, nil
}
