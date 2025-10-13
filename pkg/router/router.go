// Package router
package router

import (
	"context"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"selfier/internal/modules/job"
	"selfier/pkg/config"
	"selfier/pkg/inngest"
	"selfier/pkg/middleware"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/inngest/inngestgo"
)

// NewRouter creates and configures a new router with both HTTP API endpoints
// and the Inngest event handler.
func NewRouter(
	jobHTTPHandler job.HTTPHandler,
	jobEventHandler job.EventHandler,
	inngestCfg *config.InngestConfig,
	serverCfg *config.ServerConfig,
	baseLogger *slog.Logger,
) http.Handler {

	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("Selfier API", "1.0.0"))

	api.UseMiddleware(middleware.AuthMiddleware())
	api.UseMiddleware(middleware.RequestIDMiddleware())
	api.UseMiddleware(middleware.LoggerMiddleware(baseLogger))

	// Group all API routes under /api/v1
	apiV1 := huma.NewGroup(api, "/api/v1")

	huma.Register(apiV1, huma.Operation{
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
		resp := &struct {
			Body struct {
				Message string `json:"message"`
			}
		}{}
		resp.Body.Message = "health check ok"
		return resp, nil
	})

	huma.Register(apiV1, huma.Operation{
		OperationID:   "create-job",
		Method:        http.MethodPost,
		Path:          "/jobs",
		Summary:       "Create a new job",
		Tags:          []string{"Jobs"},
		DefaultStatus: http.StatusCreated,
	}, jobHTTPHandler.CreateJob)

	huma.Register(apiV1, huma.Operation{
		OperationID: "get-jobs-by-user",
		Method:      http.MethodGet,
		Path:        "/jobs",
		Summary:     "Get all jobs",
		Tags:        []string{"Jobs"},
	}, jobHTTPHandler.GetJobsByUser)

	huma.Register(apiV1, huma.Operation{
		OperationID: "get-job-by-id",
		Method:      http.MethodGet,
		Path:        "/jobs/{id}",
		Summary:     "Get job by id",
		Tags:        []string{"Jobs"},
	}, jobHTTPHandler.GetJobByID)

	huma.Register(apiV1, huma.Operation{
		OperationID: "delete-job-by-id",
		Method:      http.MethodDelete,
		Path:        "/jobs/{id}",
		Summary:     "Delete job by id",
		Tags:        []string{"Jobs"},
	}, jobHTTPHandler.Delete)

	huma.Register(apiV1, huma.Operation{
		OperationID: "get-presigned-url",
		Method:      http.MethodGet,
		Path:        "/jobs/{job_id}/tasks/{task_id}/images/{image_id}",
		Summary:     "Get presigned url",
		Tags:        []string{"Jobs"},
	}, jobHTTPHandler.GetPresignedURL)

	// --- Inngest Event Handler Setup ---
	client := inngest.NewClient(inngestCfg)

	_, err := inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{
			ID:   job.TaskCreatedEventTopic,
			Name: runtime.FuncForPC(reflect.ValueOf(job.EventHandler.TaskCreated).Pointer()).Name(),
		},
		inngestgo.EventTrigger(job.TaskCreatedEventTopic, nil),
		jobEventHandler.TaskCreated,
	)

	if err != nil {
		panic(err)
	}

	router.Handle("/api/inngest", client.Serve())

	corsHandler := middleware.CORSMiddleware(serverCfg.CORSAllowedOrigins)
	finalHandler := corsHandler(router)
	return finalHandler
}
