package job

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
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

	handleFailure := func(err error, msg string) (any, error) {
		log.Error(msg, slog.String("error", err.Error()))
		if _, updateErr := e.UpdateTaskStatus(ctx, taskID, StatusFailed); updateErr != nil {
			log.Error("failed to mark task as failed", slog.String("error", updateErr.Error()))
			return nil, updateErr
		}
		return nil, err
	}

	// Step 1: Update status to Running
	_, err := step.Run(ctx, "update-task-status-running", func(ctx context.Context) (*Task, error) {
		return e.UpdateTaskStatus(ctx, taskID, StatusRunning)
	})
	if err != nil {
		return handleFailure(err, "failed to update task status to running")
	}

	// Step 2: Get task
	task, err := step.Run(ctx, "get-task-by-id", func(ctx context.Context) (*Task, error) {
		return e.GetTaskByID(ctx, taskID)
	})
	if err != nil {
		return handleFailure(err, "failed to get task")
	}

	// Step 3: Get presigned url
	presignedURL, err := step.Run(ctx, "get-presigned-url", func(ctx context.Context) (string, error) {
		return e.GetPresignedURL(ctx, taskID, task.InputImage.ID)
	})
	if err != nil {
		return handleFailure(err, "failed to get presigned url")
	}

	// Step 4: Generate deselfie
	outputBase64, err := step.Run(ctx, "generate-deselfie", func(ctx context.Context) (any, error) {
		return e.GenerateDeselfie(ctx, presignedURL, GetMapFromJSON(task.Options))
	})
	if err != nil {
		return handleFailure(err, "failed to generate deselfie")
	}

	// Step 5: Upload output image
	if _, err := step.Run(ctx, "upload-output-image", func(ctx context.Context) (*Image, error) {
		outputBytes, err := base64.StdEncoding.DecodeString(outputBase64.(string))
		if err != nil {
			_, err := handleFailure(err, "failed to decode output base64")
			return nil, err
		}
		imageReader := bytes.NewReader(outputBytes)
		return e.UploadOutputImage(ctx, taskID, imageReader, fmt.Sprintf("output-%s.jpg", taskID))
	}); err != nil {
		return handleFailure(err, "failed to upload output image")
	}

	// Step 6: Update status to Finished
	if _, err := step.Run(ctx, "update-task-status-completed", func(ctx context.Context) (*Task, error) {
		return e.UpdateTaskStatus(ctx, taskID, StatusCompleted)
	}); err != nil {
		return handleFailure(err, "failed to update task status to completed")
	}

	// Step 7: Get task
	if _, err := step.Run(ctx, "get-task-by-id", func(ctx context.Context) (*Task, error) {
		return e.GetTaskByID(ctx, taskID)
	}); err != nil {
		return handleFailure(err, "failed to get task")
	}
	return true, nil
}
