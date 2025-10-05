package job

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"selfier/pkg/middleware"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type httpHandlerImpl struct {
	Service
}

func NewHandler(service Service) HTTPHandler {
	return &httpHandlerImpl{Service: service}
}

func (h *httpHandlerImpl) CreateJob(ctx context.Context, input *CreateInput) (*JobOutput, error) {
	log := middleware.GetLogger(ctx)
	formData := input.RawBody.Data()
	log.Debug("create job", slog.Any("formData", formData))

	// Validate input size
	if formData.Image.Size > 10*1024*1024 {
		log.Error("invalid image size", slog.Int("size", int(formData.Image.Size)))
		return nil, huma.Error422UnprocessableEntity("invalid image size")
	}

	// Validate readable input image
	buf, _ := io.ReadAll(formData.Image.File)
	mimeType := http.DetectContentType(buf)
	isImage := strings.HasPrefix(mimeType, "image/")
	if !isImage {
		log.Error("invalid image mime type", slog.String("mime_type", mimeType))
		return nil, huma.Error400BadRequest("invalid image mime type")
	}

	// Transform options
	var options []map[string]any
	if err := json.Unmarshal([]byte(formData.Options), &options); err != nil {
		panic(err)
	}
	if len(options) == 0 {
		return nil, huma.NewError(http.StatusBadRequest, "The 'config' array cannot be empty.")
	}

	// Send to service
	job, err := h.Service.CreateJob(ctx, options, bytes.NewReader(buf), formData.Image.Filename)
	if err != nil {
		log.Error("failed to create job", slog.String("error", err.Error()))
		return nil, huma.Error500InternalServerError("")
	}

	return &JobOutput{Body: job}, nil
}

func (h *httpHandlerImpl) GetJobByID(ctx context.Context, input *GetByIDInput) (*JobOutput, error) {

	log := middleware.GetLogger(ctx)

	job, err := h.Service.GetJobByID(ctx, input.ID)
	if err != nil {
		log.Error("job service failed to get job", slog.String("error", err.Error()))
		if err == ErrRecordNotFound {
			return nil, huma.Error404NotFound("job id " + input.ID + " not found")
		}
		return nil, huma.Error500InternalServerError("")
	}

	return &JobOutput{Body: job}, nil
}

func (h *httpHandlerImpl) GetJobsByUser(ctx context.Context, input *struct{}) (*JobsOutput, error) {

	log := middleware.GetLogger(ctx)

	jobs, err := h.Service.GetJobsByUser(ctx)
	if err != nil {
		log.Error("job service failed to get jobs by user", slog.String("error", err.Error()))
		return nil, huma.Error500InternalServerError("")
	}

	return &JobsOutput{Body: jobs}, nil
}

func (h *httpHandlerImpl) Delete(ctx context.Context, input *DeleteInput) (*struct{}, error) {

	log := middleware.GetLogger(ctx)

	err := h.DeleteJob(ctx, input.ID)
	if err != nil {
		log.Error("job service failed to delete job", slog.String("error", err.Error()))
		return nil, huma.Error500InternalServerError("")
	}

	return nil, nil
}

func (h *httpHandlerImpl) GetPresignedURL(ctx context.Context, input *GetPresignedURLInput) (*GetPresignedURLOutput, error) {

	log := middleware.GetLogger(ctx)

	presignedURL, err := h.Service.GetPresignedURL(ctx, input.TaskID, input.ImageID)
	if err != nil {
		log.Error("job service failed to create presigned url", slog.String("error", err.Error()))
		if err == ErrRecordNotFound {
			return nil, huma.Error404NotFound("job id " + input.TaskID + " not found")
		}
		return nil, huma.Error500InternalServerError("")
	}

	return &GetPresignedURLOutput{
		Body: &GetPresignedURLOutputBody{
			URL: presignedURL,
		}}, nil

}
