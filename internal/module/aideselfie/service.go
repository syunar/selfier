package aideselfie

type aideselfieServiceImpl struct {
	aideselfieProvider AIdeselfieProvider
}

func NewAideselfieService(provider AIdeselfieProvider) AIdeselfieService {
	return &aideselfieServiceImpl{aideselfieProvider: provider}
}

func (s *aideselfieServiceImpl) Generate() {
	s.aideselfieProvider.Generate()
}

func (s *aideselfieServiceImpl) GenerateMultiple() {
	s.aideselfieProvider.Generate()
}
