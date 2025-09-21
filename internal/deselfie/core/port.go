// Package core
package core

import (
	"context"
)

type DeselfiePayload struct {
	InputImageURL string `json:"input_image_url"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Prompt        string `json:"prompt"`
}

type DeselfieResult struct {
	OutputImageURL string `json:"output_image_url"`
}

type DeselfiePipeline interface {
	Run(ctx context.Context, payload DeselfiePayload) (*DeselfieResult, error)
}
