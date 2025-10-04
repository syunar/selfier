// Package event contains the ports or interefaces for the event module
package event

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"selfier/internal/module/aideselfie"
	"selfier/internal/module/job"
	"selfier/pkg/middleware"

	"github.com/inngest/inngestgo"
	"github.com/inngest/inngestgo/step"
)

type JobCreatedEvent inngestgo.GenericEvent[job.JobCreatedEventData]

type EventHandler interface {
	JobCreated(ctx context.Context, input inngestgo.Input[job.JobCreatedEventData]) (any, error)
}

type eventHandlerImpl struct {
	jobService        job.JobService
	aideselfieService aideselfie.AIdeselfieService
}

func NewEventHandler(
	jobService job.JobService,
	aideselfieService aideselfie.AIdeselfieService,
) EventHandler {
	return &eventHandlerImpl{
		jobService:        jobService,
		aideselfieService: aideselfieService,
	}
}

func (e *eventHandlerImpl) JobCreated(
	ctx context.Context,
	input inngestgo.Input[job.JobCreatedEventData],
) (any, error) {

	log := middleware.GetLogger(ctx)

	// Update Job Status to "Running"
	_, err := step.Run(ctx, "update-job-status-running", func(ctx context.Context) (bool, error) {
		_, err := e.jobService.UpdateJobStatus(ctx, input.Event.Data.JobID, job.StatusRunning)
		if err != nil {
			log.Error("job service failed to update job status", slog.String("error", err.Error()))
			return false, err
		}
		return true, nil
	})

	if err != nil {
		log.Error("job service failed to update job status", slog.String("error", err.Error()))
		_, err := e.jobService.UpdateJobStatus(ctx, input.Event.Data.JobID, job.StatusFailed)
		if err != nil {
			log.Error("job service failed to update job status", slog.String("error", err.Error()))
			return false, err
		}
		return false, err
	}

	// Get presigned url
	presignedURL, err := step.Run(
		ctx,
		"get-presigned-url",
		func(ctx context.Context) (string, error) {
			presignedURL, err := e.jobService.GetPresignedURL(
				ctx,
				input.Event.Data.JobID,
				input.Event.Data.InputImageID,
			)
			if err != nil {
				log.Error(
					"job service failed to create presigned url",
					slog.String("error", err.Error()),
				)
				return "", err
			}
			return presignedURL, nil
		},
	)

	if err != nil {
		log.Error("job service failed to create presigned url", slog.String("error", err.Error()))
		// Update Job Status
		_, err := e.jobService.UpdateJobStatus(ctx, input.Event.Data.JobID, job.StatusFailed)
		if err != nil {
			log.Error("job service failed to update job status", slog.String("error", err.Error()))
			return false, err
		}
		return false, err
	}

	// Generate
	imageBase64, err := step.Run(ctx, "generate", func(ctx context.Context) (string, error) {
		encoded, err := e.aideselfieService.Generate(
			ctx,
			presignedURL,
			input.Event.Data.ModelConfig,
		)
		if err != nil {
			log.Error("aideselfie service failed to generate", slog.String("error", err.Error()))
			return "", err
		}

		return encoded, nil
	})

	if err != nil {
		log.Error("aideselfie service failed to generate", slog.String("error", err.Error()))
		// Update Job Status
		_, err := e.jobService.UpdateJobStatus(ctx, input.Event.Data.JobID, job.StatusFailed)
		if err != nil {
			log.Error("job service failed to update job status", slog.String("error", err.Error()))
			return false, err
		}
		return false, err
	}

	// Upload image to S3
	_, err = step.Run(ctx, "upload-image-blob", func(ctx context.Context) (string, error) {
		decoded, err := base64.StdEncoding.DecodeString(imageBase64)
		if err != nil {
			log.Error("failed to decode base64", slog.String("error", err.Error()))
			return "", err
		}
		imageReader := bytes.NewReader(decoded)
		imageKey, err := e.jobService.UploadImage(ctx, imageReader, input.Event.Data.JobID)
		if err != nil {
			log.Error("job service failed to upload image", slog.String("error", err.Error()))
			return "", err
		}
		return imageKey, nil
	})

	if err != nil {
		log.Error("job service failed to upload image", slog.String("error", err.Error()))
		// Update Job Status
		_, err := e.jobService.UpdateJobStatus(ctx, input.Event.Data.JobID, job.StatusFailed)
		if err != nil {
			log.Error("job service failed to update job status", slog.String("error", err.Error()))
			return false, err
		}
		return false, err
	}

	// Update Job Status to "Finished"
	_, err = step.Run(ctx, "update-job-status-finished", func(ctx context.Context) (bool, error) {
		_, err := e.jobService.UpdateJobStatus(ctx, input.Event.Data.JobID, job.StatusFinished)
		if err != nil {
			log.Error("job service failed to update job status", slog.String("error", err.Error()))
			return false, err
		}
		return true, nil
	})

	if err != nil {
		log.Error("job service failed to update job status", slog.String("error", err.Error()))
		return false, err
	}

	return true, nil
}
