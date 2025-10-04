package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"selfier/pkg/middleware"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/inngest/inngestgo"
	"github.com/inngest/inngestgo/step"
	"github.com/inngest/inngestgo/stephttp"
)

func main() {
	provider := stephttp.Setup(stephttp.SetupOpts{
		Domain: "http://localhost:8080", // add your api domain here.
		Optional: stephttp.OptionalSetupOpts{
			BaseURL: "http://localhost:8288",
		},
	},
	)
	dev := true
	var producer, err = inngestgo.NewClient(inngestgo.ClientOpts{
		AppID: "acme-storefront-app",
		Dev:   &dev,
	})
	if err != nil {
		panic(err)
	}
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("Selfier API", "1.0.0"))
	// api.UseMiddleware(middleware.AuthMiddleware())
	// api.UseMiddleware(middleware.RequestIDMiddleware())
	api.UseMiddleware(middleware.LoggerMiddleware(slog.Default()))
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		ctx = huma.WithValue(ctx, "producer", producer)
		next(ctx)
	})

	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "get-health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Description: "Check if the API is healthy",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*struct {
		Body struct {
			Message string `json:"message"`
		}
	}, error) {

		log := middleware.GetLogger(ctx)
		producer, ok := ctx.Value("producer").(inngestgo.Client)
		if !ok {
			log.Error("failed to get producer")
			return nil, fmt.Errorf("failed to get producer")
		}

		log.Info("initial delay")
		_, err := step.Run(ctx, "initial-delay", func(ctx context.Context) (bool, error) {
			time.Sleep(2 * time.Second)
			return true, nil
		})
		if err != nil {
			log.Error("initial delay failed", slog.String("error", err.Error()))
			return nil, err
		}

		log.Info("send background job")

		result, err := producer.Send(context.Background(), inngestgo.Event{
			Name: "api/account.created",
			Data: map[string]any{
				"AccountID": "123",
			},
		},
		)

		log.Info("send background job result", slog.String("result", result))

		if err != nil {
			log.Error("send background job failed", slog.String("error", err.Error()))
			return nil, err
		}
		// _, err = step.Run(ctx, "send-welcome-email", func(ctx context.Context) (bool, error) {
		// 	time.Sleep(2 * time.Second)
		// 	return true, nil
		// })

		log.Info("health check ok")

		resp := &struct { //nolint:exhaustruct
			Body struct {
				Message string `json:"message"`
			}
		}{}
		resp.Body.Message = "health check ok"
		return resp, nil
	})

	// rest api & producer
	slog.Info("## RESTAPI AND PRODUCER ##")
	go func() {
		err := http.ListenAndServe(":8080", provider.Middleware(router))
		if err != nil {
			panic(err)
		}
	}()

	// consumer
	slog.Info("## CONSUMER ##")

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
			Timeout: time.Second * 5,
			// Match events where the user_id is the same in the async sampled event.
			If: inngestgo.StrPtr("event.data.user_id == async.data.user_id"),
		},
	)
	if err == step.ErrEventNotReceived {
		// A function wasn't created within 3 days.  Send a follow-up email.
		_, _ = step.Run(ctx, "follow-up-email", func(ctx context.Context) (any, error) {
			slog.Info("follow-up email sent")
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
