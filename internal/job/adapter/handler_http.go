// Package adapter
package adapter

import (
	"encoding/json"
	"net/http"
	deselfieCore "selfier/internal/deselfie/core"
	jobCore "selfier/internal/job/core"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type CreateJobRequest struct {
	Type    jobCore.JobType `json:"type" binding:"required"`
	Payload json.RawMessage `json:"payload" binding:"required"`
}

type JobResponse struct {
	ID      string            `json:"id"`
	Type    jobCore.JobType   `json:"type"`
	Status  jobCore.JobStatus `json:"status"`
	Payload json.RawMessage   `json:"payload"`
	Result  json.RawMessage   `json:"result,omitempty"`
	Error   *string           `json:"error,omitempty"`
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

// DeselfieHTTPHandler implements core.DeselfieHandler
type jobHTTPHandler struct {
	service jobCore.JobService
}

func NewJobHTTPHandler(service jobCore.JobService) jobCore.JobHTTPHandler {
	return &jobHTTPHandler{service: service}
}

func (h *jobHTTPHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// --- GATEWAY VALIDATION ---
	// The handler acts as a gateway, validating both the job type and its specific payload.
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
		// Validation passed for this case. Execution will continue after the switch.

	// case otherCore.JobTypeOtherJob:
	//     // Handle validation for another job type here
	//     // No break needed here either.
	//     return // if there's an error

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job type specified"})
		return // Important to return here to stop execution
	}
	// --- END VALIDATION ---

	createdJob, err := h.service.CreateJob(c.Request.Context(), req.Type, req.Payload)
	if err != nil {
		// This could be a validation error from the service or a database error.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapJobToResponse(createdJob))
}

func (h *jobHTTPHandler) GetAllJobs(c *gin.Context) {
	jobs, err := h.service.GetAllJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map the slice of domain jobs to a slice of response DTOs
	resp := make([]JobResponse, len(jobs))
	for i, job := range jobs {
		resp[i] = mapJobToResponse(job)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *jobHTTPHandler) GetJobByID(c *gin.Context) {
	id := c.Param("id")
	job, err := h.service.GetJobByID(c.Request.Context(), id)
	if err != nil {
		// A common error here would be the job not being found.
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, mapJobToResponse(job))
}

func (h *jobHTTPHandler) DeleteJobByID(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteJobByID(c.Request.Context(), id); err != nil {
		// Could be a not found error or a database error.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
