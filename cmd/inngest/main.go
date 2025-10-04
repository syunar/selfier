package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/inngest/inngestgo"
	"github.com/inngest/inngestgo/step"
)

func main() {

	slog.Info("starting server")
	dev := true
	client, err := inngestgo.NewClient(inngestgo.ClientOpts{ //nolint:exhaustruct
		AppID: "core",
		Dev:   &dev,
	})
	if err != nil {
		panic(err)
	}
	_, err = inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{ //nolint:exhaustruct
			ID:   "account-created",
			Name: "Account creation flow",
		},
		// Run on every api/account.created event.
		inngestgo.EventTrigger("api/account.created", nil),
		AccountCreated,
	)
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":8081", client.Serve()) //nolint:gosec
	if err != nil {
		panic(err)
	}

	slog.Info("server started")
}

// AccountCreated is a durable function which runs any time the "api/account.created"
// event is received by Inngest.
//
// It is invoked by Inngest, with each step being backed by Inngest's orchestrator.
// Function state is automatically managed, and persists across server restarts,
// cloud migrations, and language changes.
func AccountCreated(
	ctx context.Context,
	input inngestgo.Input[AccountCreatedEventData],
) (any, error) {
	// Sleep for a second, minute, hour, week across server restarts.
	step.Sleep(ctx, "initial-delay", time.Second)
	// Run a step which emails the user.  This automatically retries on error.
	// This returns the fully typed result of the lambda.
	result, err := step.Run(ctx, "on-user-created", func(ctx context.Context) (bool, error) {
		// Run any code inside a step.
		// result, err := emails.Send(emails.Opts{})
		slog.Info("email sent")
		return true, nil
	})
	if err != nil {
		// This step retried 5 times by default and permanently failed.
		return nil, err
	}
	// `result` is  fully typed from the lambda
	_ = result

	// Sample from the event stream for new events.  The function will stop
	// running and automatially resume when a matching event is found, or if
	// the timeout is reached.
	fn, err := step.WaitForEvent[FunctionCreatedEvent](
		ctx,
		"wait-for-activity",
		step.WaitForEventOpts{
			Name:    "Wait for a function to be created",
			Event:   "api/function.created",
			Timeout: time.Hour * 72,
			// Match events where the user_id is the same in the async sampled event.
			If: inngestgo.StrPtr("event.data.user_id == async.data.user_id"),
		},
	)
	if err == step.ErrEventNotReceived {
		// A function wasn't created within 3 days.  Send a follow-up email.
		_, _ = step.Run(ctx, "follow-up-email", func(ctx context.Context) (any, error) {
			// ...
			return true, nil
		})
		return nil, nil
	}

	// The event returned from `step.WaitForEvent` is fully typed.
	fmt.Println(fn.Data.FunctionID)

	return nil, nil
}

// AccountCreatedEvent represents the fully defined event received when an account is created.
//
// This is shorthand for defining a new Inngest-conforming struct:
//
//	type AccountCreatedEvent struct {
//		Name      string                  `json:"name"`
//		Data      AccountCreatedEventData `json:"data"`
//		User      map[string]any          `json:"user"`
//		Timestamp int64                   `json:"ts,omitempty"`
//		Version   string                  `json:"v,omitempty"`
//	}
type AccountCreatedEvent inngestgo.GenericEvent[AccountCreatedEventData]
type AccountCreatedEventData struct {
	AccountID string
}

type FunctionCreatedEvent inngestgo.GenericEvent[FunctionCreatedEventData]
type FunctionCreatedEventData struct {
	FunctionID string
}
