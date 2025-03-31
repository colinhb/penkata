package penkata

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
)

// WeightTransform represents different methods to transform raw count weights
type WeightTransform int

const (
	// Raw keeps the original count values
	Raw WeightTransform = iota
	// Log1p applies log1p to reduce impact of common bigrams
	Log1p
	// Normal converts counts to probabilities (count/total)
	Normal
)

// BigramWeights stores the weights calculated from the TSV data
type BigramWeights struct {
	Weights   map[string]float64
	Transform WeightTransform
	Total     int
}

// LoadBigramWeights reads and processes the TSV file with the specified weight transformation
func LoadBigramWeights(filepath string, transform WeightTransform) (*BigramWeights, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	weights := make(map[string]float64)
	total := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		bigram := parts[0]
		count, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}

		// Store raw counts for now
		weights[bigram] = float64(count)
		total += count
	}

	switch transform {
	case Log1p:
		for bigram, count := range weights {
			weights[bigram] = math.Log1p(count)
		}
	case Normal:
		totalFloat := float64(total)
		for bigram, count := range weights {
			weights[bigram] = count / totalFloat
		}
	case Raw:
		// Keep original counts
	}

	return &BigramWeights{
		Weights:   weights,
		Transform: transform,
		Total:     total,
	}, nil
}
