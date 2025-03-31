package penkata

import (
	"fmt"
)

// WindowParams holds configuration parameters for window processing
type WindowParams struct {
	ID       string // Unique identifier
	Weights  *BigramWeights
	MaxChars int
}

// NewWindowParams creates a new parameter set
func NewWindowParams(weights *BigramWeights, maxChars int) *WindowParams {
	// Use the maxChars value as the ID for simpler identification
	id := fmt.Sprintf("%d", maxChars)

	return &WindowParams{
		ID:       id,
		Weights:  weights,
		MaxChars: maxChars,
	}
}
