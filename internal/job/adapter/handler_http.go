// Package adapter
package adapter

import (
	"encoding/json"
	"net/http"

	// These are your internal packages, make sure the paths are correct
	deselfieCore "selfier/internal/deselfie/core"
	jobCore "selfier/internal/job/core"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type CreateJobRequest struct {
	Type    jobCore.JobType `json:"type" binding:"required" enums:"deselfie" example:"deselfie"`
	Payload json.RawMessage `json:"payload" binding:"required" example:"{\"image_url\":\"https://example.com/image.jpg\"}" swaggertype:"string"`
}

type JobResponse struct {
	ID      string            `json:"id" example:"job-12345"`
	Type    jobCore.JobType   `json:"type" example:"deselfie"`
	Status  jobCore.JobStatus `json:"status" example:"pending"`
	Payload json.RawMessage   `json:"payload" example:"{\"image_url\":\"https://example.com/image.jpg\"}" swaggertype:"string"`
	Result  json.RawMessage   `json:"result,omitempty" example:"{\"processed_url\":\"https://example.com/result.jpg\"}" swaggertype:"string"`
	Error   *string           `json:"error,omitempty" example:"an error occurred"`
}

// GinError represents a generic error response for Gin.
type GinError struct {
	Error string `json:"error"`
}

func mapJobToResponse(job *jobCore.Job) JobResponse {
	return JobResponse{
		ID:      job.ID,
		Type:    job.Type,
		Status:  job.Status,
		Payload: job.Payload,
		Result:  job.Result,
		Error:   job.Error,
	}
}

// jobHTTPHandler implements core.JobHTTPHandler
type jobHTTPHandler struct {
	service jobCore.JobService
}

func NewJobHTTPHandler(service jobCore.JobService) jobCore.JobHTTPHandler {
	return &jobHTTPHandler{service: service}
}

// CreateJob godoc
// @Summary      Create a new job
// @Description  Creates a new job. The structure of the 'payload' depends on the 'type'.
// @Description  For 'deselfie', the payload must be `{"image_url": "string"}`.
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Param        job  body      CreateJobRequest  true  "Create Job Request"
// @Success      201  {object}  JobResponse
// @Failure      400  {object}  GinError "Invalid request body or payload"
// @Failure      500  {object}  GinError "Internal server error"
// @Router       /jobs [post]
func (h *jobHTTPHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// --- GATEWAY VALIDATION ---
	switch req.Type {
	case jobCore.JobTypeDeselfie:
		var deselfiePayload deselfieCore.DeselfiePayload
		if err := json.Unmarshal(req.Payload, &deselfiePayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deselfie payload format: " + err.Error()})
			return
		}
		if err := binding.Validator.ValidateStruct(deselfiePayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payload validation failed: " + err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job type specified"})
		return
	}
	// --- END VALIDATION ---

	createdJob, err := h.service.CreateJob(c.Request.Context(), req.Type, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapJobToResponse(createdJob))
}

// GetAllJobs godoc
// @Summary      Get all jobs
// @Description  Retrieves a list of all jobs.
// @Tags         jobs
// @Produce      json
// @Success      200  {array}   JobResponse
// @Failure      500  {object}  GinError "Internal server error"
// @Router       /jobs [get]
func (h *jobHTTPHandler) GetAllJobs(c *gin.Context) {
	jobs, err := h.service.GetAllJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]JobResponse, len(jobs))
	for i, job := range jobs {
		resp[i] = mapJobToResponse(job)
	}

	c.JSON(http.StatusOK, resp)
}

// GetJobByID godoc
// @Summary      Get a job by ID
// @Description  Retrieves a single job by its unique ID.
// @Tags         jobs
// @Produce      json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  JobResponse
// @Failure      404  {object}  GinError "Job not found"
// @Router       /jobs/{id} [get]
func (h *jobHTTPHandler) GetJobByID(c *gin.Context) {
	id := c.Param("id")
	job, err := h.service.GetJobByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, mapJobToResponse(job))
}

// DeleteJobByID godoc
// @Summary      Delete a job by ID
// @Description  Deletes a job by its unique ID.
// @Tags         jobs
// @Produce      json
// @Param        id   path      string  true  "Job ID"
// @Success      204  "No Content"
// @Failure      500  {object}  GinError "Internal server error"
// @Router       /jobs/{id} [delete]
func (h *jobHTTPHandler) DeleteJobByID(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteJobByID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
