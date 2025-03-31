// Package penkata provides text analysis and window-based processing capabilities
package penkata

import "strings"

// Window represents a section of text with metadata.
// It maintains both the original words and their derived bigrams for scoring.
type Window struct {
	words   []string       // Raw words, not normalized
	bigrams map[string]int // Map of bigrams to their counts for scoring
	params  *WindowParams  // Configuration parameters for this window
}

// NewWindow creates a new empty window with the provided parameters.
func NewWindow(params *WindowParams) Window {
	return Window{
		words:   []string{},
		bigrams: make(map[string]int),
		params:  params,
	}
}

// Text returns the window content as a space-joined string.
func (w *Window) Text() string {
	return strings.Join(w.words, " ")
}

// Score calculates the window score based on weighted unique bigrams.
//
// NOTE: Due to floating-point arithmetic not being associative (a+(b+c) ≠ (a+b)+c),
// and Go's non-deterministic map iteration order, this function may return slightly
// different scores for identical bigram sets across different runs. Even tiny
// floating-point differences (e.g., 1287.91 vs 1287.9100000000001) can cause
// the passage selection logic to choose different passages despite having the
// "same" score when formatted. For deterministic results, consider:
//  1. Using a sorted slice of bigrams to ensure consistent addition order
//  2. Implementing fixed-precision decimal arithmetic
//  3. Using epsilon comparisons for score equality checks
func (w *Window) Score() float64 {
	score := 0.0
	for bigram, count := range w.bigrams {
		if count > 0 {
			if weight, exists := w.params.Weights.Weights[bigram]; exists {
				score += weight
			}
		}
	}
	return score
}

// pushWord adds a word to the window and updates the bigram set.
// This method modifies the window in place.
func (w *Window) pushWord(word string) {
	// Add the word to words slice
	w.words = append(w.words, word)

	// Initialize bigrams map if needed
	if w.bigrams == nil {
		w.bigrams = make(map[string]int)
	}

	// Get all bigrams from this word
	bigrams := extractBigramsFromWord(word)
	if len(bigrams) == 0 {
		return
	}

	// Add all extracted bigrams to the set
	for _, bg := range bigrams {
		w.bigrams[bg]++
	}
}

// shiftWord removes the first word from the window and updates the bigram counts accordingly.
// It returns the removed word or an empty string if the window was empty.
// This method modifies the window in place.
func (w *Window) shiftWord() string {
	// Check if there are any words to shift
	if len(w.words) == 0 {
		return ""
	}

	// Get the first word
	word := w.words[0]

	// Remove the first word from the slice
	w.words = w.words[1:]

	// Ensure bigrams map is initialized
	if w.bigrams == nil {
		w.bigrams = make(map[string]int)
		return word
	}

	// Get all bigrams from the removed word
	bigrams := extractBigramsFromWord(word)
	if len(bigrams) == 0 {
		return word
	}

	// Decrement the counter for each bigram from the word
	for _, bg := range bigrams {
		w.bigrams[bg]--

		// If the count reaches 0, remove the bigram from the map
		if w.bigrams[bg] == 0 {
			delete(w.bigrams, bg)
		}
	}

	return word
}

// Clone creates and returns a deep copy of the window.
func (w *Window) Clone() Window {
	newWindow := Window{
		words:   make([]string, len(w.words)),
		bigrams: make(map[string]int),
		params:  w.params, // Copy the params pointer
	}

	// Copy words
	copy(newWindow.words, w.words)

	// Copy bigrams
	for bg, count := range w.bigrams {
		newWindow.bigrams[bg] = count
	}

	return newWindow
}

// Size returns the total character count of the window, including spaces between words.
func (w *Window) Size() int {
	if len(w.words) == 0 {
		return 0
	}

	totalSize := 0
	for _, word := range w.words {
		totalSize += len([]rune(word))
	}

	// Add spaces between words
	totalSize += len(w.words) - 1

	return totalSize
}

// Words returns a copy of the window's words slice.
func (w *Window) Words() []string {
	result := make([]string, len(w.words))
	copy(result, w.words)
	return result
}

// Bigrams returns a copy of the window's bigrams map.
func (w *Window) Bigrams() map[string]int {
	result := make(map[string]int, len(w.bigrams))
	for bg, count := range w.bigrams {
		result[bg] = count
	}
	return result
}

// IsZero returns true if the Window is a zero value or effectively empty.
func (w *Window) IsZero() bool {
	// Note: by definition len(nil) == 0 for slices and maps
	return len(w.words) == 0 && len(w.bigrams) == 0
}

// MaybeAddWord attempts to add a word to a copy of the window if it fits within maxChars.
// Returns the new window and true if successful, or an empty window and false if the word wouldn't fit.
func (w *Window) MaybeAddWord(word string, params *WindowParams) (Window, bool) {
	// Check for nil params or empty word
	if params == nil {
		return Window{}, false
	}

	// Check if word would fit
	newSize := w.Size()
	if len(w.words) > 0 {
		newSize++ // Add space before new word
	}
	newSize += len([]rune(word))

	if newSize > params.MaxChars {
		return Window{}, false // Word doesn't fit
	}

	// Word fits, create new window with word added
	newWindow := w.Clone()
	newWindow.pushWord(word)
	return newWindow, true
}

// AddWord adds a word to the window and ensures the window doesn't exceed the maximum
// character limit by removing words from the beginning if necessary.
// This method modifies the window in place.
func (w *Window) AddWord(word string) {
	if w.params == nil {
		return // Cannot add word without params defining max size
	}

	// Add the new word
	w.pushWord(word)

	// Remove words from the beginning until we're under the size limit
	for w.Size() > w.params.MaxChars {
		if w.shiftWord() == "" {
			break // Safety check in case window is empty
		}
	}
}

// Params returns the window's parameters.
func (w *Window) Params() *WindowParams {
	return w.params
}
