package provider

import (
	"llm-gateway/internal/config"
	"net/http"
	"sync"
)

// Manager holds the configuration and clients for all downstream providers.
type Manager struct {
	providers map[string]*http.Client
	configs   map[string]config.Provider
	mu        sync.RWMutex
}

// NewManager creates and returns a new provider manager.
func NewManager(providers []config.Provider) *Manager {
	m := &Manager{
		providers: make(map[string]*http.Client),
		configs:   make(map[string]config.Provider),
	}

	for _, p := range providers {
		if p.Enabled {
			m.configs[p.Name] = p
			m.providers[p.Name] = &http.Client{
				Timeout: p.Timeout,
			}
		}
	}

	return m
}

// GetClient returns the http.Client for a given provider.
func (m *Manager) GetClient(providerName string) *http.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[providerName]
}

// GetConfig returns the configuration for a given provider.
func (m *Manager) GetConfig(providerName string) (config.Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, ok := m.configs[providerName]
	return cfg, ok
}

// GetAllProviderConfigs returns a slice of all provider configurations.
func (m *Manager) GetAllProviderConfigs() []config.Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	configs := make([]config.Provider, 0, len(m.configs))
	for _, cfg := range m.configs {
		configs = append(configs, cfg)
	}
	return configs
}
