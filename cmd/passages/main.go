package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"github.com/colinhb/penkata/pkg/penkata"
)

// flagValue provides a simple reusable implementation of flag.Value interface
type flagValue struct {
	stringFunc func() string
	setFunc    func(string) error
}

func (v *flagValue) String() string {
	return v.stringFunc()
}

func (v *flagValue) Set(value string) error {
	return v.setFunc(value)
}

// Config holds program configuration from command-line flags
type Config struct {
	BigramFile       string                    // Path to bigram weights file
	DirPath          string                    // Root directory to search
	MaxChars         []int                     // Multiple max passage lengths
	TopN             int                       // Number of passages to display
	OutputFile       string                    // Path to output file (optional)
	Verbose          bool                      // Whether to print intermediate results
	WeightTransforms []penkata.WeightTransform // Types of transformations to apply to weights
}

// parseFlags processes command-line arguments and validates required parameters
func parseFlags() *Config {
	config := &Config{}

	// Create temporary slices for flag parsing
	var sizes []int
	var weightTransforms []penkata.WeightTransform

	// Create wrapper values that implement flag.Value
	sizeValue := flagValue{
		stringFunc: func() string { return fmt.Sprintf("%v", sizes) },
		setFunc: func(value string) error {
			intVal, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			sizes = append(sizes, intVal)
			return nil
		},
	}

	transformValue := flagValue{
		stringFunc: func() string { return fmt.Sprintf("%v", weightTransforms) },
		setFunc: func(value string) error {
			var transform penkata.WeightTransform
			switch value {
			case "raw":
				transform = penkata.Raw
			case "log1p":
				transform = penkata.Log1p
			case "normal":
				transform = penkata.Normal
			default:
				return fmt.Errorf("invalid weight transformation type: %s. Must be one of: raw, log1p, normal", value)
			}
			weightTransforms = append(weightTransforms, transform)
			return nil
		},
	}

	flag.StringVar(&config.BigramFile, "f", "sonnets-bigrams.tsv", "TSV file with bigram counts")
	flag.StringVar(&config.DirPath, "d", "", "Directory to walk for text files")
	flag.Var(&sizeValue, "c", "Maximum characters in passage (can be specified multiple times: -c 150 -c 300)")
	flag.IntVar(&config.TopN, "n", 50, "Number of top-scoring passages to display per size")
	flag.StringVar(&config.OutputFile, "o", "", "Output file for results (optional)")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose output of intermediate results")
	flag.Var(&transformValue, "w", "Weight transformation type (raw, log1p, normal) (can be specified multiple times: -w log1p -w normal)")
	flag.Parse()

	if config.DirPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -d <directory> [-f <bigram-file>] [-c <size>] [-n <n>] [-o <output-file>] [-v] [-w <weight-transform>]\n", os.Args[0])
		os.Exit(1)
	}

	// Default to 200 if no sizes specified
	if len(sizes) == 0 {
		config.MaxChars = []int{200}
	} else {
		config.MaxChars = sizes
	}

	// Default to raw if no weight transforms specified
	if len(weightTransforms) == 0 {
		config.WeightTransforms = []penkata.WeightTransform{penkata.Raw}
	} else {
		config.WeightTransforms = weightTransforms
	}

	return config
}

// insertSorted adds a passage to the results list while maintaining sorted order (highest score first)
// and keeping only the top N results
func insertSorted(results []penkata.Passage, newPassage penkata.Passage, maxResults int) []penkata.Passage {
	// Find the insertion position (where new score is higher than existing)
	pos := len(results)
	for i, p := range results {
		if newPassage.Score() > p.Score() {
			pos = i
			break
		}
	}

	// If the position is beyond our max results, it doesn't make the cut
	if pos >= maxResults {
		return results
	}

	// Insert the new result
	if len(results) < maxResults {
		// We have room to append
		if pos == len(results) {
			return append(results, newPassage)
		}
		results = append(results, penkata.Passage{}) // Make room
	}

	// Shift elements to make room for insertion
	copy(results[pos+1:], results[pos:])
	results[pos] = newPassage

	// Trim if necessary
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// getTransformName returns the string representation of a weight transform
func getTransformName(transform penkata.WeightTransform) string {
	switch transform {
	case penkata.Log1p:
		return "log1p"
	case penkata.Normal:
		return "normal"
	default:
		return "raw"
	}
}

// printResults outputs all passages from the map to the specified writer
func printResults(w io.Writer, bestPassagesByParams map[*penkata.WindowParams][]penkata.Passage, paramsList []*penkata.WindowParams, separateBySection bool) {
	if separateBySection {
		// Print each section with its own header and results
		for _, params := range paramsList {
			passages := bestPassagesByParams[params]
			if len(passages) == 0 {
				continue
			}

			transformName := getTransformName(params.Weights.Transform)

			// Print section header
			fmt.Fprintf(w, "\n=== Results for %d character passages (%s) ===\n",
				params.MaxChars, transformName)

			// Print TSV header for this section
			fmt.Fprintln(w, "transform\tmaxChar\tpath\tscore\tsize\ttext")

			// Print each passage
			for _, p := range passages {
				fmt.Fprintf(w, "%s\t%d\t%s\t%.2f\t%d\t%s\n",
					transformName, params.MaxChars, p.FilePath, p.Score(), p.Size(), p.Text())
			}
		}
	} else {
		// Print a single header followed by all results without section headers
		fmt.Fprintln(w, "transform\tmaxChar\tpath\tscore\tsize\ttext")

		// Process each parameter set in the original order
		for _, params := range paramsList {
			passages := bestPassagesByParams[params]
			if len(passages) == 0 {
				continue
			}

			transformName := getTransformName(params.Weights.Transform)

			// Print each passage
			for _, p := range passages {
				fmt.Fprintf(w, "%s\t%d\t%s\t%.2f\t%d\t%s\n",
					transformName, params.MaxChars, p.FilePath, p.Score(), p.Size(), p.Text())
			}
		}
	}
}

func main() {
	// Parse command line arguments into a configuration structure
	config := parseFlags()

	// Check if output file is writable if specified
	var outputFile *os.File
	if config.OutputFile != "" {
		var err error
		outputFile, err = os.Create(config.OutputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
	}

	// Create window parameters for each combination of size and weight transform
	var paramsList []*penkata.WindowParams
	for _, transform := range config.WeightTransforms {
		// Load bigram weights with the current transformation
		weights, err := penkata.LoadBigramWeights(config.BigramFile, transform)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading bigrams with transform %v: %v\n", transform, err)
			os.Exit(1)
		}

		// Create parameters for each size with this weight transform
		for _, size := range config.MaxChars {
			params := penkata.NewWindowParams(weights, size)
			// No longer need to create and store the ID
			paramsList = append(paramsList, params)
		}
	}

	// Set up concurrent processing based on available CPU cores
	numWorkers := runtime.NumCPU() * 3 / 2 // Use 1.5x CPU cores for workers

	// Channel sizing is proportional to worker count to balance memory usage and throughput
	errCh := make(chan error, 2*numWorkers) // Small buffer for errors (rare event)

	// Synchronized error collection - errors are collected and displayed immediately
	// but also tracked for final exit status
	var errs []error
	var errMu sync.Mutex

	// Start error reporting goroutine - runs throughout program execution
	// and prints errors as they occur while maintaining a full record
	go func() {
		for err := range errCh {
			errMu.Lock()
			errs = append(errs, err)
			errMu.Unlock()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}()

	// Channel for distributing file paths to worker goroutines
	// Larger buffer (10x workers) helps keep workers busy during uneven file processing times
	filesCh := make(chan string, 10*numWorkers)

	// Start file discovery goroutine to walk directory and feed work queue
	go func() {
		defer close(filesCh) // Signal workers when no more files are coming
		err := filepath.WalkDir(config.DirPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				errCh <- fmt.Errorf("access error for %s: %w", path, err)
				return nil // Continue walking despite errors
			}
			if d.IsDir() {
				return nil // Skip directories
			}
			if penkata.HasTextExtension(path) {
				filesCh <- path // Send text files to workers
			}
			return nil
		})
		if err != nil {
			errCh <- err // Report errors in the walk itself
		}
	}()

	var wg sync.WaitGroup

	// Channel for workers to return passages back to main goroutine
	resultsCh := make(chan []penkata.Passage, 4*numWorkers)

	// Start worker pool - each worker processes files until channel closes
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filesCh {
				// Find the best passages in each file (one for each size)
				passages, err := penkata.FindBestPassagesInFile(path, paramsList)
				if err != nil {
					errCh <- fmt.Errorf("processing %s: %w", path, err)
					continue
				}
				resultsCh <- passages
			}
		}()
	}

	// Wait for all workers to complete, then close result and error channels
	// This signals the main goroutine that all processing is done
	go func() {
		wg.Wait()
		close(resultsCh) // No more results coming
		close(errCh)     // No more errors coming
	}()

	// Maintain sorted collections for each size
	bestPassagesByParams := make(map[*penkata.WindowParams][]penkata.Passage)
	statsByParams := make(map[*penkata.WindowParams]*penkata.Stats)

	for _, params := range paramsList {
		bestPassagesByParams[params] = make([]penkata.Passage, 0, config.TopN)
		statsByParams[params] = penkata.NewStats()
	}

	// Show initial stats header if verbose
	if config.Verbose {
		fmt.Fprintln(os.Stderr, "\nProcessing files and collecting statistics...")
	}

	// Process results as they arrive
	for passages := range resultsCh {
		for _, passage := range passages {
			// Extract params from the Window embedded in Passage
			params := passage.Window.Params()

			bestPassagesByParams[params] = insertSorted(bestPassagesByParams[params], passage, config.TopN)
			statsByParams[params].Update(passage.Score())

			// Print new best passages when found
			if bestPassagesByParams[params][0].FilePath == passage.FilePath {
				// Print the new best passage
				fmt.Fprintf(os.Stderr, "New best passage for %d characters (%s): %s\n",
					params.MaxChars, getTransformName(params.Weights.Transform), passage.Text())
			}

			if config.Verbose {
				fileCount := statsByParams[params].FilesProcessed
				if fileCount%10 == 0 {
					transformName := getTransformName(params.Weights.Transform)
					fmt.Fprintf(os.Stderr, "MaxChars %d; Weight %s: %s\n",
						params.MaxChars, transformName, statsByParams[params])
				}
			}
		}
	}

	// Final stats output if verbose
	if config.Verbose {
		fmt.Fprintln(os.Stderr, "\nFinal Statistics:")
		for _, params := range paramsList {
			transformName := getTransformName(params.Weights.Transform)
			fmt.Fprintf(os.Stderr, "MaxChars %d; Weight %s: %s\n",
				params.MaxChars, transformName, statsByParams[params])
		}
		fmt.Fprintln(os.Stderr)
	}

	// Output results to the appropriate destination
	outputDest := os.Stdout
	if outputFile != nil {
		outputDest = outputFile
	}

	// Print all results with a single function call
	printResults(outputDest, bestPassagesByParams, paramsList, outputFile == nil)

	// Exit with error status if any errors occurred during processing
	errMu.Lock()
	hasErrors := len(errs) > 0
	errMu.Unlock()

	if hasErrors {
		os.Exit(1)
	}
}
