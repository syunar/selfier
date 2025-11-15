package job

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"selfier/pkg/middleware"

	"github.com/inngest/inngestgo"
	"github.com/inngest/inngestgo/step"
)

type eventHandlerImpl struct {
	Service
}

func NewEventHandlerIngest(service Service) EventHandler {
	return &eventHandlerImpl{Service: service}
}

func (e *eventHandlerImpl) TaskCreated(ctx context.Context, input inngestgo.Input[TaskCreatedEventData]) (any, error) {
	log := middleware.GetLogger(ctx)
	taskID := input.Event.Data.TaskID

	defer func() {
		if r := recover(); r != nil {
			log.Error("recovered from panic in TaskCreated handler", "panic", r)
			_, _ = e.UpdateTaskStatus(ctx, taskID, StatusFailed)
		}
	}()

	// 1. Update status to Running.
	_, err := step.Run(ctx, "update-task-status-running", func(ctx context.Context) (*Task, error) {
		return e.UpdateTaskStatus(ctx, taskID, StatusRunning)
	})
	if err != nil {
		log.Error("failed to update task status to running", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	// 2. Get task details.
	task, err := step.Run(ctx, "get-task-by-id", func(ctx context.Context) (*Task, error) {
		return e.GetTaskByID(ctx, taskID)
	})
	if err != nil {
		log.Error("failed to get task", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	// 3. Get a presigned URL for the input image.
	presignedURL, err := step.Run(ctx, "get-presigned-url", func(ctx context.Context) (string, error) {
		return e.GetPresignedURL(ctx, taskID, task.InputImage.ID)
	})
	if err != nil {
		log.Error("failed to get presigned url", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	// 4. Generate the deselfie image (returns a base64 string).
	outputBase64, err := step.Run(ctx, "generate-deselfie", func(ctx context.Context) (string, error) {
		return e.GenerateDeselfie(ctx, presignedURL, GetMapFromJSON(task.Options))
	})
	if err != nil {
		log.Error("failed to generate deselfie", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	// Decode the base64 string. This is not a step, just a simple transformation.
	outputBytes, err := base64.StdEncoding.DecodeString(outputBase64)
	if err != nil {
		log.Error("failed to decode output base64", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	// 5. Upload the resulting image bytes.
	_, err = step.Run(ctx, "upload-output-image", func(ctx context.Context) (*Image, error) {
		imageReader := bytes.NewReader(outputBytes)
		filename := fmt.Sprintf("output-%s.jpg", taskID)
		return e.UploadOutputImage(ctx, taskID, imageReader, filename)
	})
	if err != nil {
		log.Error("failed to upload output image", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	// 6. Update status to Completed and return the final task object.
	completedTask, err := step.Run(ctx, "update-task-status-completed", func(ctx context.Context) (*Task, error) {
		return e.UpdateTaskStatus(ctx, taskID, StatusCompleted)
	})
	if err != nil {
		log.Error("failed to update task status to completed", "error", err)
		return nil, e.markTaskAsFailed(ctx, taskID, err)
	}

	log.Info("task completed successfully", "taskID", taskID)
	return completedTask, nil
}

func (e *eventHandlerImpl) markTaskAsFailed(ctx context.Context, taskID string, originalErr error) error {
	log := middleware.GetLogger(ctx)
	if _, updateErr := e.UpdateTaskStatus(ctx, taskID, StatusFailed); updateErr != nil {
		log.Error("critical: failed to mark task as failed after another error occurred",
			"taskID", taskID,
			"originalError", originalErr.Error(),
			"updateError", updateErr.Error(),
		)
	}
	return originalErr
}
