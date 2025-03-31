package penkata

import (
	"bufio"
	"os"
)

// Passage represents a Window from a specific file
type Passage struct {
	Window          // Embed Window
	FilePath string // Source file path
}

// FindBestPassagesInFile finds the best passage for each window size in a file
func FindBestPassagesInFile(filepath string, paramsList []*WindowParams) ([]Passage, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create slices for current windows and best windows
	windows := make([]Window, len(paramsList))
	bestWindows := make([]Window, len(paramsList))

	// Initialize windows for each parameter set
	for i, params := range paramsList {
		windows[i] = NewWindow(params)
		bestWindows[i] = NewWindow(params)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()

		// Update each window with the new word
		for i := range windows {
			windows[i].AddWord(word)

			// Update best window if score improved
			if windows[i].Score() > bestWindows[i].Score() {
				bestWindows[i] = windows[i].Clone()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert each best window to a passage
	result := make([]Passage, len(bestWindows))
	for i, window := range bestWindows {
		result[i] = Passage{
			Window:   window,
			FilePath: filepath,
		}
	}

	return result, nil
}
