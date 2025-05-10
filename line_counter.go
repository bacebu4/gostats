package main

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

type lineCounter struct {
	mu    sync.Mutex
	cache map[string]int
}

func (c *lineCounter) count(path string) (int, error) {
	c.mu.Lock()
	if value, exists := c.cache[path]; exists {
		c.mu.Unlock()
		DPrintf("Got from cache!")
		return value, nil
	}
	c.mu.Unlock()

	content, err := os.ReadFile(path)

	if err != nil {
		return 0, fmt.Errorf("cannot read file for counting lines: %w", err)
	}

	result := bytes.Count(content, []byte{'\n'})

	// If the file doesn't end with a newline, we need to add 1
	// This handles the case of the last line not ending with a newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		result++
	}

	c.mu.Lock()
	c.cache[path] = result
	c.mu.Unlock()

	return result, nil
}

func makeLineCounter() *lineCounter {
	result := &lineCounter{}

	result.cache = make(map[string]int)

	return result
}
