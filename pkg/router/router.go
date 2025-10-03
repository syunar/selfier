// Package router
package router

import (
	"context"
	"log/slog"
	"net/http"
	"selfier/internal/module/job"
	"selfier/pkg/middleware"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func NewRouter(jobHandler job.JobHTTPHandler, baseLogger *slog.Logger) *http.ServeMux {
	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("Selfier API", "1.0.0"))

	api.UseMiddleware(middleware.AuthMiddleware())
	api.UseMiddleware(middleware.RequestIDMiddleware())
	api.UseMiddleware(middleware.LoggerMiddleware(baseLogger))

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "get-health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Description: "Check if the API is healthy",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*struct {
		Body struct {
			Message string `json:"message"`
		}
	}, error) {
		resp := &struct { //nolint:exhaustruct
			Body struct {
				Message string `json:"message"`
			}
		}{}
		resp.Body.Message = "health check ok"
		return resp, nil
	})

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID:   "create-job",
		Method:        http.MethodPost,
		Path:          "/jobs",
		Summary:       "Create a new job",
		Tags:          []string{"Jobs"},
		DefaultStatus: http.StatusCreated,
	}, jobHandler.CreateJob)

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "get-jobs",
		Method:      http.MethodGet,
		Path:        "/jobs",
		Summary:     "Get all jobs",
		Tags:        []string{"Jobs"},
	}, jobHandler.GetJobs)

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "get-job-by-id",
		Method:      http.MethodGet,
		Path:        "/jobs/{id}",
		Summary:     "Get job by id",
		Tags:        []string{"Jobs"},
	}, jobHandler.GetJobByID)

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "delete-job-by-id",
		Method:      http.MethodDelete,
		Path:        "/jobs/{id}",
		Summary:     "Delete job by id",
		Tags:        []string{"Jobs"},
	}, jobHandler.DeleteJobByID)

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "get-job-images-by-id",
		Method:      http.MethodGet,
		Path:        "/jobs/{id}/images",
		Summary:     "Get job images by id",
		Tags:        []string{"Jobs"},
	}, jobHandler.GetJobImagesByID)

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "get-presigned-url",
		Method:      http.MethodGet,
		Path:        "/jobs/{job_id}/images/{image_id}",
		Summary:     "Get presigned url",
		Tags:        []string{"Jobs"},
	}, jobHandler.GetPresignedURL)

	return router
}
