package job

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"selfier/pkg/middleware"
)

type deselfiePipelineAPI struct {
}

func NewDeselfiePipelineAPI() DeselfiePipeline {
	return &deselfiePipelineAPI{}
}

func (p *deselfiePipelineAPI) Generate(ctx context.Context, imageURL string, options map[string]any) (string, error) {

	log := middleware.GetLogger(ctx)

	// download image
	resp, err := http.Get(imageURL)
	if err != nil {
		log.Error("failed to download image", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	defer func() { _ = resp.Body.Close() }()
	// read image bytes
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read image bytes", slog.String("error", err.Error()))
		return "", ErrInternal
	}

	mockOutputBase64 := base64.StdEncoding.EncodeToString(imageBytes)

	return mockOutputBase64, nil
}
