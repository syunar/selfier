package job

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"selfier/pkg/middleware"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type jobHTTPHandlerImpl struct {
	jobService JobService
}

func NewJobHTTPHandler(jobService JobService) JobHTTPHandler {
	return &jobHTTPHandlerImpl{
		jobService: jobService,
	}
}

func (h *jobHTTPHandlerImpl) CreateJob(ctx context.Context, input *CreateJobInput) (*JobOutput, error) {
	log := middleware.GetLogger(ctx)

	formData := input.RawBody.Data()

	// 1. Validate the input
	// Validate the model config
	var modelConfig map[string]interface{}
	if err := json.Unmarshal([]byte(formData.ModelConfig), &modelConfig); err != nil {
		log.Error("cannot unmarshal model modelConfig", slog.String("error", err.Error()))
		return nil, huma.Error422UnprocessableEntity("invalid model config")
	}

	// formData.Image.Filename should be .jpg or .png or .webp or .jpeg
	ext := strings.ToLower(filepath.Ext(formData.Image.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		// valid
	default:
		log.Error("invalid image filename", slog.String("filename", formData.Image.Filename))
		return nil, huma.Error422UnprocessableEntity("invalid image filename")
	}

	// formData.Image.Size should be less than 10MB
	if formData.Image.Size > 10*1024*1024 {
		log.Error("invalid image size", slog.Int("size", int(formData.Image.Size)))
		return nil, huma.Error422UnprocessableEntity("invalid image size")
	}

	// 2. Read the uploaded image file into a buffer.
	// This ensures the io.Reader is fully consumed and can be passed to the service.
	// Using a buffer allows the service to potentially re-read the data if needed (e.g., for hashing/uploading).
	buf, _ := io.ReadAll(formData.Image.File)
	mimeType := http.DetectContentType(buf)
	isImage := strings.HasPrefix(mimeType, "image/")
	if !isImage {
		log.Error("invalid image mime type", slog.String("mime_type", mimeType))
		return nil, huma.Error400BadRequest("invalid image mime type")
	}

	// 3. Delegate to the JobService, passing the necessary parsed data
	jobBody, err := h.jobService.CreateJob(
		ctx,
		formData.Type,
		modelConfig,
		bytes.NewReader(buf),
		formData.Image.Filename,
	)
	if err != nil {
		log.Error("job service failed to create job", slog.String("error", err.Error()))
		return nil, huma.Error500InternalServerError("")
	}
	log.Info("job created", slog.String("job_id", jobBody.ID))
	// 4. Wrap the service's result in the handler's output structure
	return &JobOutput{Body: *jobBody}, nil
}

func (h *jobHTTPHandlerImpl) GetJobs(ctx context.Context, input *struct{}) (*JobsOutput, error) {

	log := middleware.GetLogger(ctx)

	jobs, err := h.jobService.GetJobs(ctx)
	if err != nil {
		log.Error("job service failed to get jobs", slog.String("error", err.Error()))
		return nil, huma.Error500InternalServerError("")
	}

	return &JobsOutput{
		Body: jobs,
	}, nil
}

func (h *jobHTTPHandlerImpl) GetJobByID(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*JobOutput, error) {

	log := middleware.GetLogger(ctx)

	job, err := h.jobService.GetJobByID(ctx, input.ID)
	if err != nil {
		log.Error("job service failed to get job", slog.String("error", err.Error()))
		if err == ErrNotFound {
			return nil, huma.Error404NotFound("job id " + input.ID + " not found")
		}
		return nil, huma.Error500InternalServerError("")
	}

	resp := &JobOutput{
		Body: *job,
	}

	return resp, nil
}

func (h *jobHTTPHandlerImpl) DeleteJobByID(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*struct{}, error) {

	log := middleware.GetLogger(ctx)

	err := h.jobService.DeleteJobByID(ctx, input.ID)
	if err != nil {
		log.Error("job service failed to delete job", slog.String("error", err.Error()))
		return nil, huma.Error500InternalServerError("")
	}

	return &struct{}{}, nil
}

func (h *jobHTTPHandlerImpl) GetJobResultsByID(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*JobResultsOutput, error) {

	log := middleware.GetLogger(ctx)

	jobResults, err := h.jobService.GetJobResultsByID(ctx, input.ID)
	if err != nil {
		log.Error("job service failed to get job results", slog.String("error", err.Error()))
		if err == ErrNotFound {
			return nil, huma.Error404NotFound("job id " + input.ID + " not found")
		}
		return nil, huma.Error500InternalServerError("")
	}

	resp := &JobResultsOutput{
		Body: jobResults,
	}
	return resp, nil
}
