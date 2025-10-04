package aideselfie

import (
	"context"
	"log/slog"
	"selfier/pkg/middleware"
)

type AIdeselfieService interface {
	Generate(ctx context.Context, imageURL string, options map[string]interface{}) (string, error)
}

type aideselfieServiceImpl struct {
	aideselfieProvider AIdeselfieProvider
}

func NewAideselfieService(provider AIdeselfieProvider) AIdeselfieService {
	return &aideselfieServiceImpl{aideselfieProvider: provider}
}

func (s *aideselfieServiceImpl) Generate(
	ctx context.Context,
	imageURL string,
	options map[string]interface{},
) (string, error) {
	log := middleware.GetLogger(ctx)
	encoded, err := s.aideselfieProvider.Generate(ctx, imageURL, options)
	if err != nil {
		log.Error("aideselfie provider failed to generate", slog.String("error", err.Error()))
		return "", err
	}

	return encoded, err
}
