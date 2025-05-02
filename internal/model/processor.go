package model

import (
	"context"
	"encoding/json"

	"github.com/google/generative-ai-go/genai"

	"github.com/renniemaharaj/news/internal/types"
	"github.com/renniemaharaj/news/internal/validation"

	"github.com/renniemaharaj/news/pkg/pool"
	"github.com/renniemaharaj/news/pkg/transformer/gemi"
)

// Constructs an input for transformer communication
func getInput(content []string) gemi.Input {
	contentBytes, err := json.Marshal(content)
	var HISTORY = []*genai.Content{}
	if err == nil {
		return gemi.Input{
			Current: genai.Text(string(contentBytes)),
			History: HISTORY,
			Context: []map[string]string{},
		}
	}

	return gemi.Input{}
}

const (
	queues  = 2
	backoff = 2
)

// Prompt function interfaces with transformer package on our behalf
func Prompt(content []string) (types.Wrapper, error) {
	p := pool.Instance{}
	p.InitializePool()

	// call the transformer package, queued, exponential backoff and validation
	resp, err := p.QueuedEVS(context.Background(), getInput(content), validation.Validate, queues, backoff)
	if err != nil {
		return types.Wrapper{}, err
	}

	var reports types.Wrapper

	// queuedEVS already handles validation into
	err = json.Unmarshal([]byte(resp), &reports)
	if err != nil {
		return types.Wrapper{}, err
	}

	return reports, nil
}
