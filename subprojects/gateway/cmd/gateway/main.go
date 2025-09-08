package main

import (
	"fmt"
	"llm-gateway/internal/config"
	"llm-gateway/internal/core"
	coremw "llm-gateway/internal/core/middleware"
	"llm-gateway/internal/core/provider"
	"llm-gateway/internal/core/router"
	"llm-gateway/internal/logging"
	"llm-gateway/internal/ratelimit"
	"llm-gateway/internal/transport/handlers"
	transportmw "llm-gateway/internal/transport/middleware"
	"net/http"
	"time"
)

func main() {
	// 1. Load Configuration
	logger := logging.NewLogger()
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Components
	providerManager := provider.NewManager(cfg.Providers)
	modelsCache := core.NewModelsCache()
	coreRouter := router.NewRouter(cfg.Strategies)

	// 2a. Setup Core Middleware (Post-Forwarding)
	responseMiddleware := coremw.ElasticCompletionLogger
	proxy := core.NewProxy(providerManager, coreRouter, responseMiddleware)

	// 3. Start the Model Fetcher
	modelFetcher := core.NewModelFetcher(providerManager, modelsCache, 10*time.Minute)
	modelFetcher.Start()
	defer modelFetcher.Stop()

	// 4. Setup HTTP Server
	gatewayHandler := handlers.NewGatewayHandler(modelsCache, proxy)
	mux := http.NewServeMux()
	gatewayHandler.RegisterRoutes(mux)

	// 4a. Setup Transport Middleware (Pre-Forwarding)
	transportMiddlewareManager := transportmw.NewManager(logger)

	var middlewares []transportmw.Middleware
	middlewares = append(middlewares, transportMiddlewareManager.Logging)

	// Initialize OIDC Authenticator if enabled
	if cfg.Auth.Enabled {
		auth := transportmw.NewOIDCAuthenticator(logger, cfg.Auth.Issuer, cfg.Auth.Audience, cfg.Auth.CacheTTL)
		middlewares = append(middlewares, transportMiddlewareManager.Authentication(auth))
		logger.Info("OIDC authentication enabled")

		// Add the Authorization middleware right after Authentication
		authz := transportmw.NewAuthorizer(logger, cfg.Providers, modelsCache)
		middlewares = append(middlewares, transportMiddlewareManager.Authorization(authz))
		logger.Info("Model authorization enabled")
	}

	// Initialize Rate Limiter if enabled
	if cfg.RateLimit.Enabled {
		var store ratelimit.RateLimiterStore
		switch cfg.RateLimit.Backend {
		case "redis":
			store = ratelimit.NewRedisStore(cfg.RateLimit.RedisAddress)
			logger.Info("Rate limiting enabled with redis backend")
		case "memory":
			store = ratelimit.NewMemoryStore()
			logger.Info("Rate limiting enabled with memory backend")
		default:
			logger.Warnf("Unknown rate limit backend '%s', defaulting to in-memory", cfg.RateLimit.Backend)
			store = ratelimit.NewMemoryStore()
		}
		middlewares = append(middlewares, transportMiddlewareManager.RateLimiter(store, cfg.RateLimit))
	}

	chainedHandler := transportmw.Chain(middlewares...)(mux)

	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Starting server on %s", serverAddr)

	// 5. Start the Server
	if err := http.ListenAndServe(serverAddr, chainedHandler); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}