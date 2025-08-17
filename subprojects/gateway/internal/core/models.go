package core

import (
	"sync"
)

// Model represents a single LLM model returned by the gateway.
// It includes the provider name for namespacing.
type Model struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Created  int64  `json:"created"`
	OwnedBy  string `json:"owned_by"`
	Provider string `json:"provider"`
}

// ProviderModel is a temporary struct to unmarshal the response from a downstream provider.
// It does not include the provider name, as that's added during aggregation.
type ProviderModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ProviderModelsList is a temporary struct to unmarshal the list response from a provider.
type ProviderModelsList struct {
	Data []ProviderModel `json:"data"`
}


// ModelsCache holds the aggregated list of models from all providers.
type ModelsCache struct {
	models map[string][]Model
	mu     sync.RWMutex
}

// NewModelsCache creates a new model cache.
func NewModelsCache() *ModelsCache {
	return &ModelsCache{
		models: make(map[string][]Model),
	}
}

// SetModels sets the models for a given provider.
func (c *ModelsCache) SetModels(providerName string, models []Model) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.models[providerName] = models
}

// GetAllModels returns a flattened list of all models from all providers.
func (c *ModelsCache) GetAllModels() []Model {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allModels []Model
	for _, providerModels := range c.models {
		allModels = append(allModels, providerModels...)
	}
	return allModels
}
