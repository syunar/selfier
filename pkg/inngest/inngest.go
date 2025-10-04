// Package inngest contains the ports or interefaces for the inngest module
package inngest

import (
	"context"
	"selfier/internal/module/event"
	"selfier/internal/module/job"
	"selfier/pkg/config"

	"github.com/inngest/inngestgo"
)

func NewClient(cfg *config.InngestConfig) inngestgo.Client {

	//nolint:exhaustruct
	client, err := inngestgo.NewClient(inngestgo.ClientOpts{
		AppID: cfg.AppID,
		Dev:   &cfg.Dev,
	})

	if err != nil {
		panic(err)
	}

	return client
}

func NewConsumerClient(
	cfg *config.InngestConfig,
	eventHandler event.EventHandler,
) inngestgo.Client {
	client := NewClient(cfg)

	_, err := inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{
			ID:   "job-created",
			Name: "Job creation flow",
		},
		inngestgo.EventTrigger(job.JobCreatedEventTopic, nil),
		func(ctx context.Context, input inngestgo.Input[job.JobCreatedEventData]) (any, error) {
			return eventHandler.JobCreated(ctx, input)
		},
	)

	if err != nil {
		panic(err)
	}

	return client
}
