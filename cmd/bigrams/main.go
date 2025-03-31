package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	fun "github.com/colinhb/penkata/pkg/myfuncs"
	a "github.com/colinhb/penkata/pkg/mytypes"
	penkata "github.com/colinhb/penkata/pkg/penkata"
)

func main() {
	dirFlag := flag.String("d", "", "directory to process")
	flag.Parse()
	if *dirFlag == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -d <directory>\n", os.Args[0])
		os.Exit(1)
	}

	errCh := make(chan error, 100)
	defer close(errCh)

	// Track if we encountered any errors
	var errs []error
	var errMu sync.Mutex

	// Start error reporting in background
	go func() {
		for err := range errCh {
			errMu.Lock()
			errs = append(errs, err)
			errMu.Unlock()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}()

	filesCh := make(chan string, 100)
	go func() {
		defer close(filesCh)
		err := filepath.WalkDir(*dirFlag, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				errCh <- fmt.Errorf("access error for %s: %w", path, err)
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if penkata.HasTextExtension(path) {
				filesCh <- path
			}
			return nil
		})
		if err != nil {
			errCh <- err
		}
	}()

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup

	resultsCh := make(chan map[string]int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		// This is our worker - captures filesCh and errCh from closure
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filesCh {
				counts, err := penkata.CountBigramsInFile(path)
				if err != nil {
					errCh <- fmt.Errorf("processing %s: %w", path, err)
					continue
				}
				resultsCh <- counts
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Merge results from all workers.
	totals := make(map[string]int)
	for res := range resultsCh {
		fun.MergeMaps(totals, res)
	}

	// Convert map to sorted slice using the generic Tuple type
	results := make([]a.Tuple[string, int], 0, len(totals))
	for k, v := range totals {
		results = append(results, a.Tuple[string, int]{First: k, Second: v})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Second > results[j].Second
	})
	for _, r := range results {
		fmt.Printf("%s\t%d\n", r.First, r.Second)
	}

	errMu.Lock()
	hasErrors := fun.Ternary(len(errs) > 0, true, false)
	errMu.Unlock()

	if hasErrors {
		os.Exit(1)
	}
}
