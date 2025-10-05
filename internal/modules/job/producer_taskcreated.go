package job

import (
	"context"
	"encoding/json"
	"log/slog"
	"selfier/pkg/middleware"

	"github.com/google/uuid"
	"github.com/inngest/inngestgo"
)

type taskCreatedProducerInngest struct {
	inngestCleint inngestgo.Client
}

func NewTaskCreatedProducerInngest(producer inngestgo.Client) TaskCreatedProducer {
	return &taskCreatedProducerInngest{producer}
}

func (p *taskCreatedProducerInngest) Send(ctx context.Context, event *TaskCreatedEventData) (string, error) {

	log := middleware.GetLogger(ctx)

	eventMap := make(map[string]any)

	// struct to json
	data, err := json.Marshal(event)
	if err != nil {
		log.Error("failed to marshal event", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	// json to map
	if err := json.Unmarshal(data, &eventMap); err != nil {
		log.Error("failed to unmarshal event", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	return p.inngestCleint.Send(ctx, inngestgo.Event{
		ID:   func(s string) *string { return &s }(uuid.New().String()),
		Name: TaskCreatedEventTopic,
		Data: eventMap,
	})
}
