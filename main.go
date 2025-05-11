package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"runtime"
	"sync"

	glob "github.com/bmatcuk/doublestar/v4"
	gitignore "github.com/sabhiram/go-gitignore"
)

const Debug = false

func DPrintf(format string, a ...any) {
	if Debug {
		fmt.Printf(format, a...)
	}
}

type Config struct {
	TargetPatterns []string `json:"targetPatterns"`
	TotalPatterns  []string `json:"totalPatterns"`
}

type fileKind string

const (
	targetFile fileKind = "target"
	totalFile  fileKind = "total"
)

type pattern struct {
	kind  fileKind
	value string
}

const CONFIG_NAME = ".gostats.json"

func readConfigPatterns(homeDir string, workingDir string) ([]pattern, error) {
	marshaledConfig, err := os.ReadFile(fmt.Sprintf("%s/%s", workingDir, CONFIG_NAME))

	if errors.Is(err, os.ErrNotExist) {
		marshaledConfig, err = os.ReadFile(fmt.Sprintf("%s/%s", homeDir, CONFIG_NAME))

		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New(".gostats.json was not found")
		} else if err != nil {
			return nil, fmt.Errorf("unexpected error on reading .gostats.json config: %w", err)
		}
	}

	var config Config

	if err = json.Unmarshal(marshaledConfig, &config); err != nil {
		return nil, fmt.Errorf("cannot unmarshal .gostats.json config: %w", err)
	}

	DPrintf("Target patterns: %v\n", config.TargetPatterns)
	DPrintf("Total patterns: %v\n", config.TotalPatterns)

	var result []pattern

	for _, patternValue := range config.TargetPatterns {
		result = append(result, pattern{kind: targetFile, value: patternValue})
	}

	for _, patternValue := range config.TotalPatterns {
		result = append(result, pattern{kind: totalFile, value: patternValue})
	}

	return result, nil
}

func readGitignore() (*gitignore.GitIgnore, error) {
	gitIgnorePattern, err := gitignore.CompileIgnoreFileAndLines(".gitignore", ".git")

	if errors.Is(err, os.ErrNotExist) {
		gitIgnorePattern = gitignore.CompileIgnoreLines(".git")
		return gitIgnorePattern, nil
	} else if err != nil {
		return nil, fmt.Errorf("cannot parse git ignore: %w", err)
	}

	return gitIgnorePattern, nil
}

type path struct {
	value string
	kind  fileKind
}

func findPathsByPatterns(patterns []pattern, gitIgnorePattern *gitignore.GitIgnore, pathJobs chan<- path, workingDir string) {
	defer func() {
		close(pathJobs)
	}()

	fileSystem := os.DirFS(workingDir)
	err := fs.WalkDir(fileSystem, ".", func(pathValue string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", pathValue, err)
			return nil
		}

		DPrintf("Checking path: %v\n", pathValue)

		if d.IsDir() {
			if gitIgnorePattern.MatchesPath(pathValue) {
				return fs.SkipDir
			}
			return nil
		}

		if gitIgnorePattern.MatchesPath(pathValue) {
			return nil
		}

		for _, pattern := range patterns {
			matched, err := glob.Match(pattern.value, pathValue)

			if err != nil {
				return fmt.Errorf("error on matching: %w", err)
			}

			if matched {
				pathJobs <- path{value: pathValue, kind: pattern.kind}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("cannot walk working directory: %v\n", err)
	}
}

func worker(pathJobs <-chan path, results chan<- result, errors chan<- error, wg *sync.WaitGroup, lineCounter *lineCounter) {
	defer wg.Done()

	for path := range pathJobs {
		resultValue, err := lineCounter.count(path.value)

		if err != nil {
			errors <- fmt.Errorf("error counting lines in %s: %w", path, err)
			continue
		}

		results <- result{kind: path.kind, value: resultValue, path: path.value}
	}
}

type result struct {
	kind  fileKind
	value int
	path  string
}

func countLinesByPatterns(patterns []pattern, gitIgnorePattern *gitignore.GitIgnore, lineCounter *lineCounter, workingDir string) (map[fileKind]int, error) {
	pathJobs := make(chan path, 200)

	go findPathsByPatterns(patterns, gitIgnorePattern, pathJobs, workingDir)

	results := make(chan result, 200)
	errors := make(chan error, 10)

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(pathJobs, results, errors, &wg, lineCounter)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	sumByKind := make(map[fileKind]int)

	// Patterns might intersect each other within group
	// Deduplicate found path by each kind
	alreadyCounted := make(map[fileKind](map[string]bool))
	alreadyCounted[targetFile] = make(map[string]bool)
	alreadyCounted[totalFile] = make(map[string]bool)

	for result := range results {
		if alreadyCounted[result.kind][result.path] {
			continue
		}
		sumByKind[result.kind] += result.value
		alreadyCounted[result.kind][result.path] = true
	}

	for err := range errors {
		fmt.Println(err)
	}

	return sumByKind, nil
}

func main() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Printf("Cannon get user home dir: %v\n", err)
		return
	}

	workingDir, err := os.Getwd()

	if err != nil {
		fmt.Printf("Cannon get working dir: %v\n", err)
		return
	}

	patterns, err := readConfigPatterns(homeDir, workingDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	lineCounter := makeLineCounter()

	gitIgnorePattern, err := readGitignore()
	if err != nil {
		fmt.Println(err)
		return
	}

	sumByKind, err := countLinesByPatterns(patterns, gitIgnorePattern, lineCounter, workingDir)

	if err != nil {
		fmt.Printf("Error counting: %v\n", err)
		return
	}

	sumTarget := sumByKind[targetFile]
	sumTotal := sumByKind[totalFile]

	var result float64
	if sumTotal == 0 {
		result = 0
	} else {
		result = math.Round(float64(sumTarget)/float64(sumTotal)*10000) / 100
	}

	fmt.Printf("Sum target: %v LOC\n", sumTarget)
	fmt.Printf("Sum total: %v LOC\n", sumTotal)
	fmt.Printf("Percentage: %v%%\n", result)
}
