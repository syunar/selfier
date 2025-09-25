// Package aideselfie contains the ports or interefaces for the aideselfie module
package aideselfie

type AIdeselfieService interface {
	Generate()
	GenerateMultiple()
}

type AIdeselfieProvider interface {
	Generate()
}
