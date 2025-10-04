package job

import (
	"context"
	"encoding/json"
	"log/slog"
	"selfier/pkg/middleware"

	"github.com/google/uuid"
	"github.com/inngest/inngestgo"
)

type JobCreatedEventData struct {
	JobID        string         `json:"job_id"         example:"123"`
	InputImageID string         `json:"input_image_id" example:"abc"`
	ModelConfig  map[string]any `json:"model_config"   example:"{\"prompt\": \"a photo of a person\"}"`
}

type JobEventPublisher interface {
	JobCreated(ctx context.Context, event *JobCreatedEventData) (string, error)
}

type jobEventPublisherInngest struct {
	producer inngestgo.Client
}

func NewJobEventPublisherInngest(producer inngestgo.Client) JobEventPublisher {
	return &jobEventPublisherInngest{producer}
}

func (j *jobEventPublisherInngest) JobCreated(
	ctx context.Context,
	event *JobCreatedEventData,
) (string, error) {

	log := middleware.GetLogger(ctx)

	b, err := json.Marshal(event)
	if err != nil {
		log.Error("failed to marshal event", slog.String("error", err.Error()))
		return "", err
	}
	var eventMap map[string]any
	if err := json.Unmarshal(b, &eventMap); err != nil {
		log.Error("failed to unmarshal event", slog.String("error", err.Error()))
		return "", err
	}

	return j.producer.Send(ctx, inngestgo.Event{
		ID:   func(s string) *string { return &s }(uuid.New().String()),
		Name: JobCreatedEventTopic,
		Data: eventMap,
	})
}
