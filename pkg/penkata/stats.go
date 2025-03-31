package penkata

import (
	"fmt"
	"math"
)

// Stats tracks statistics about processed passages incrementally
type Stats struct {
	FilesProcessed int     // Count of files processed
	MaxScore       float64 // Highest score seen
	MinScore       float64 // Lowest score seen
	Sum            float64 // Sum of all scores
	Mean           float64 // Running mean
	M2             float64 // Running sum of squares of differences from mean
}

// NewStats creates a new Stats object with initialized values
func NewStats() *Stats {
	return &Stats{
		MaxScore: -1.0,            // Start with negative value so any valid score will be higher
		MinScore: math.MaxFloat64, // Start with maximum value so any valid score will be lower
	}
}

// Update updates stats with a new score using Welford's online algorithm
func (s *Stats) Update(score float64) {
	s.FilesProcessed++

	// Update max and min scores
	if score > s.MaxScore {
		s.MaxScore = score
	}
	if score < s.MinScore {
		s.MinScore = score
	}

	// Update sum
	s.Sum += score

	// Welford's online algorithm for variance
	delta := score - s.Mean
	s.Mean += delta / float64(s.FilesProcessed)
	delta2 := score - s.Mean
	s.M2 += delta * delta2
}

// StdDev returns the current standard deviation
func (s *Stats) StdDev() float64 {
	if s.FilesProcessed < 2 {
		return 0
	}
	variance := s.M2 / float64(s.FilesProcessed)
	return math.Sqrt(variance)
}

// String returns a formatted string representation of current stats
func (s *Stats) String() string {
	return fmt.Sprintf("Files: %d | Max Score: %.2f | Min Score: %.2f | Mean: %.2f | StdDev: %.2f",
		s.FilesProcessed, s.MaxScore, s.MinScore, s.Mean, s.StdDev())
}
