package core

import (
	"encoding/json"
	"fmt"
	"io"
	"llm-gateway/internal/config"
	"llm-gateway/internal/core/provider"
	"log"
	"net/http"
	"time"
)

// ModelFetcher is responsible for periodically fetching models from all providers.
type ModelFetcher struct {
	providerManager *provider.Manager
	modelsCache     *ModelsCache
	ticker          *time.Ticker
	stopChan        chan bool
}

// NewModelFetcher creates a new model fetcher.
func NewModelFetcher(pm *provider.Manager, mc *ModelsCache, refreshInterval time.Duration) *ModelFetcher {
	return &ModelFetcher{
		providerManager: pm,
		modelsCache:     mc,
		ticker:          time.NewTicker(refreshInterval),
		stopChan:        make(chan bool),
	}
}

// Start begins the periodic fetching of models.
func (mf *ModelFetcher) Start() {
	log.Println("Starting model fetcher...")
	// Fetch immediately on start
	mf.fetchAllModels()

	go func() {
		for {
			select {
			case <-mf.ticker.C:
				mf.fetchAllModels()
			case <-mf.stopChan:
				mf.ticker.Stop()
				return
			}
		}
	}()
}

// Stop halts the periodic fetching.
func (mf *ModelFetcher) Stop() {
	mf.stopChan <- true
}

func (mf *ModelFetcher) fetchAllModels() {
	log.Println("Fetching models from all providers...")
	providers := mf.providerManager.GetAllProviderConfigs()
	for _, p := range providers {
		if p.Enabled {
			go mf.fetchModelsForProvider(p)
		}
	}
}

func (mf *ModelFetcher) fetchModelsForProvider(p config.Provider) {
	client := mf.providerManager.GetClient(p.Name)
	if client == nil {
		log.Printf("Failed to get client for provider: %s", p.Name)
		return
	}

	req, err := http.NewRequest("GET", p.TargetURL+"/v1/models", nil)
	if err != nil {
		log.Printf("Error creating request for provider %s: %v", p.Name, err)
		return
	}

	// Add API key if it exists
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching models from provider %s: %v", p.Name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Provider %s returned non-200 status: %d", p.Name, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body from provider %s: %v", p.Name, err)
		return
	}

	var providerModelsList ProviderModelsList
	if err := json.Unmarshal(body, &providerModelsList); err != nil {
		log.Printf("Error unmarshaling models from provider %s: %v", p.Name, err)
		return
	}

	// Namespace the models and update the cache
	namespacedModels := make([]Model, len(providerModelsList.Data))
	for i, m := range providerModelsList.Data {
		namespacedModels[i] = Model{
			ID:       fmt.Sprintf("%s/%s", p.Name, m.ID),
			Object:   m.Object,
			Created:  m.Created,
			OwnedBy:  m.OwnedBy,
			Provider: p.Name,
		}
	}

	mf.modelsCache.SetModels(p.Name, namespacedModels)
	log.Printf("Successfully fetched and updated %d models for provider: %s", len(namespacedModels), p.Name)
}
