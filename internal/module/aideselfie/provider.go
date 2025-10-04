package aideselfie

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"selfier/pkg/middleware"
)

type AIdeselfieProvider interface {
	Generate(ctx context.Context, imageURL string, options map[string]interface{}) (string, error)
}

type aideselfieProviderModal struct {
}

func NewAideselfieProviderModal() AIdeselfieProvider {
	return &aideselfieProviderModal{}
}

func (a *aideselfieProviderModal) Generate(
	ctx context.Context,
	imageURL string,
	options map[string]interface{},
) (string, error) {

	log := middleware.GetLogger(ctx)

	// download image
	resp, err := http.Get(imageURL)
	if err != nil {
		log.Error("failed to download image", slog.String("error", err.Error()))
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	// read image bytes
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read image bytes", slog.String("error", err.Error()))
		return "", err
	}

	// from bytes to base64
	encoded := base64.StdEncoding.EncodeToString(imageBytes)

	return encoded, nil
}
