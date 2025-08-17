package router

import (
	"encoding/json"
	"errors"
	"io"
	"llm-gateway/internal/config"
	"strings"
)

// RequestBody is a partial representation of the incoming request body
// used to extract the model name.
type RequestBody struct {
	Model string `json:"model"`
}

// Router determines the provider strategy for a given request.
type Router struct {
strategies map[string]*config.Strategy
}

// NewRouter creates a new router.
func NewRouter(strategies []config.Strategy) *Router {
	r := &Router{
		strategies: make(map[string]*config.Strategy),
	}
	for i, s := range strategies {
		r.strategies[s.Name] = &strategies[i]
	}
	return r
}

// SelectStrategy selects a provider strategy based on the model name in the request body.
// For now, it uses a simple logic of matching the provider name from the model string.
// e.g. "openai/gpt-4o" -> "openai" strategy.
// This will be expanded to use the configured strategies.
func (r *Router) SelectStrategy(body io.Reader) (*config.Strategy, error) {
	var reqBody RequestBody
	if err := json.NewDecoder(body).Decode(&reqBody); err != nil {
		return nil, err
	}

	if reqBody.Model == "" {
		return nil, errors.New("model not found in request body")
	}

	// Simple strategy: find a strategy that has the provider from the model name.
	// This is a placeholder for the full strategy selection logic.
	parts := strings.Split(reqBody.Model, "/")
	if len(parts) < 2 {
		return nil, errors.New("invalid model format; expected 'provider/model_name'")
	}
	providerName := parts[0]

	// A real implementation would look up the strategy from the config.
	// For now, we create a dummy strategy on the fly.
	// This will be replaced with a lookup in r.strategies.
	strategy := &config.Strategy{
		Name:      "dynamic",
		Providers: []string{providerName},
	}

	return strategy, nil
}
