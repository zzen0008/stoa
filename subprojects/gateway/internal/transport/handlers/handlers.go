package handlers

import (
	"encoding/json"
	"llm-gateway/internal/core"
	"net/http"
)

// GatewayHandler holds the dependencies for the HTTP handlers.
type GatewayHandler struct {
	modelsCache *core.ModelsCache
	proxy       *core.Proxy
}

// NewGatewayHandler creates a new gateway handler.
func NewGatewayHandler(mc *core.ModelsCache, p *core.Proxy) *GatewayHandler {
	return &GatewayHandler{
		modelsCache: mc,
		proxy:       p,
	}
}

// RegisterRoutes registers the API routes.
func (h *GatewayHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/models", h.GetModels)
	mux.HandleFunc("/v1/chat/completions", h.ChatCompletions)
	mux.HandleFunc("/v1/info", h.GetInfo)
}

// GetModels handles the /v1/models endpoint.
func (h *GatewayHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	models := h.modelsCache.GetAllModels()
	// Wrap the models in a "data" object to match the OpenAI API format.
	response := struct {
		Object string       `json:"object"`
		Data   []core.Model `json:"data"`
	}{
		Object: "list",
		Data:   models,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChatCompletions handles the /v1/chat/completions endpoint.
func (h *GatewayHandler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	h.proxy.ServeHTTP(w, r)
}

// GetInfo handles the /v1/info endpoint.
func (h *GatewayHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}{
		Status:  "ok",
		Version: "0.1.0", // This could be dynamic later
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
