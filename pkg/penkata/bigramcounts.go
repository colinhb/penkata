package penkata

import (
	"bufio"
	"fmt"
	"os"

	fun "github.com/colinhb/penkata/pkg/myfuncs"
)

// countBigrams processes a file and extracts bigrams.
func CountBigramsInFile(path string) (map[string]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", path, err)
	}
	defer file.Close()

	counts := make(map[string]int)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		fun.MergeMaps(counts, countBigramsInWord(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %s: %w", path, err)
	}

	return counts, nil
}

// countBigramsInWord counts bigrams in a word.
func countBigramsInWord(word string) map[string]int {
	counts := make(map[string]int)

	// Get all bigrams from this word
	bigrams := extractBigramsFromWord(word)

	// Count each bigram
	for _, bg := range bigrams {
		counts[bg]++
	}

	return counts
}
