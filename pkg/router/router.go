// Package router
package router

import (
	"net/http"
	jobCore "selfier/internal/job/core"

	"github.com/gin-gonic/gin"
)

func NewRouter(jobHandler jobCore.JobHTTPHandler) *gin.Engine {
	// Create a new Gin router with default middleware (logger and recovery).
	r := gin.Default()

	// A simple health check endpoint to verify that the service is running.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Group all API routes under a versioned prefix for better maintainability.
	apiV1 := r.Group("/api/v1")
	{
		// Group all job-related endpoints under the "/jobs" resource.
		jobs := apiV1.Group("/jobs")
		{
			// POST /api/v1/jobs
			// Creates a new job. The request body must specify the job `type` and `payload`.
			jobs.POST("", jobHandler.CreateJob)

			// GET /api/v1/jobs
			// Retrieves a list of all jobs.
			jobs.GET("", jobHandler.GetAllJobs)

			// GET /api/v1/jobs/:id
			// Retrieves a specific job by its ID.
			jobs.GET("/:id", jobHandler.GetJobByID)

			// DELETE /api/v1/jobs/:id
			// Deletes a specific job by its ID.
			jobs.DELETE("/:id", jobHandler.DeleteJobByID)
		}
	}

	return r
}
