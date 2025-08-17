package main

import (
	"fmt"
	"llm-gateway/internal/config"
	"llm-gateway/internal/core"
	"llm-gateway/internal/core/provider"
	"llm-gateway/internal/core/router"
	"llm-gateway/internal/transport/handlers"
	"log"
	"net/http"
	"time"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Components
	providerManager := provider.NewManager(cfg.Providers)
	modelsCache := core.NewModelsCache()
	coreRouter := router.NewRouter(cfg.Strategies)
	proxy := core.NewProxy(providerManager, coreRouter)

	// 3. Start the Model Fetcher
	// Refresh models every 10 minutes.
	modelFetcher := core.NewModelFetcher(providerManager, modelsCache, 10*time.Minute)
	modelFetcher.Start()
	defer modelFetcher.Stop()

	// 4. Setup HTTP Server
	gatewayHandler := handlers.NewGatewayHandler(modelsCache, proxy)
	mux := http.NewServeMux()
	gatewayHandler.RegisterRoutes(mux)

	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)

	// 5. Start the Server
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
