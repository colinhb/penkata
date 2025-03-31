package penkata

import (
	"path/filepath"
	"strings"

	a "github.com/colinhb/penkata/pkg/mytypes"
)

func dropTerminalPunctuation(w []rune) []rune {
	if len(w) == 0 {
		return w
	}
	if strings.ContainsRune(".,;:!?", rune(w[len(w)-1])) {
		return w[:len(w)-1]
	}
	return w
}

// isApostrophe checks if a rune is any of the common apostrophe variants
func isApostrophe(r rune) bool {
	s := a.NewSet[rune]()
	s.AddAll(
		'\'',     // ASCII apostrophe (U+0027)
		'\u2019', // Right single quotation mark (’)
		'\u2018', // Left single quotation mark (')
		'\u02BC', // Modifier letter apostrophe (ʼ)
		'\u2032', // Prime (′)
	)

	return s.Contains(r)
}

// collapse (remove) common apostrophes
func collapseCommonApostraphes(w []rune) []rune {
	// Need at least 2 characters for a contraction
	if len(w) < 2 {
		return w
	}

	// Define sets of suffixes for different apostrophe positions
	s := a.NewSet[rune]().AddAll('m', 's', 'd', 't')

	// Check for apostrophe in penultimate position (len-2)
	// Handles: 'm, 's, 'd, 't
	if len(w) >= 2 && isApostrophe(w[len(w)-2]) {
		lastChar := w[len(w)-1]
		if s.Contains(lastChar) {
			return append(w[:len(w)-2], lastChar)
		}
	}

	t := a.NewSet[string]().AddAll("re", "ve", "ll")

	// Check for apostrophe in pre-penultimate position (len-3)
	// Handles: 're, 've, 'll
	if len(w) >= 4 && isApostrophe(w[len(w)-3]) {
		suffix := string(w[len(w)-2:])
		if t.Contains(suffix) {
			return append(w[:len(w)-3], w[len(w)-2:]...)
		}
	}

	return w
}

func normalizeWord(w string) string {
	wordMaps := make([](func([]rune) []rune), 0)
	wordMaps = append(
		wordMaps,
		dropTerminalPunctuation,
		collapseCommonApostraphes,
	)

	runes := []rune(w)

	// Apply word maps
	for _, fn := range wordMaps {
		runes = fn(runes)
	}

	return string(runes)
}

// extractBigramsFromWord returns all bigrams from a normalized word
func extractBigramsFromWord(word string) []string {
	normalized := normalizeWord(word)
	runes := []rune(normalized)

	if len(runes) == 0 {
		return nil
	}

	// Pre-allocate for all possible bigrams (leading + internal + trailing)
	bigrams := make([]string, 0, len(runes)+1)

	// Leading boundary bigram
	bigrams = append(bigrams, "_"+string(runes[0]))

	// Internal bigrams
	for i := 0; i < len(runes)-1; i++ {
		bigrams = append(bigrams, string(runes[i:i+2]))
	}

	// Trailing boundary bigram
	bigrams = append(bigrams, string(runes[len(runes)-1])+"_")

	return bigrams
}

func HasTextExtension(path string) bool {
	// whitelist of text file extensions
	textExts := map[string]bool{
		".txt": true,
		".md":  true,
	}

	ext := strings.ToLower(filepath.Ext(path))
	return textExts[ext]
}
