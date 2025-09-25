package aideselfie

import "fmt"

type aideselfieProviderModal struct {
}

func NewAideselfieProviderModal() AIdeselfieProvider {
	return &aideselfieProviderModal{}
}

func (a *aideselfieProviderModal) Generate() {
	// TODO: implement
	fmt.Print("aideselfie generate")
}
