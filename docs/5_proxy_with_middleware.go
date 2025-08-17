package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// This example demonstrates how to use middleware with a reverse proxy.
// Middleware is a function that wraps an HTTP handler, allowing you to run code
// before and/or after the main handler. It's the standard way to implement
// cross-cutting concerns like logging, authentication, rate limiting, etc.

// --- Backend Server ---
func backendServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Backend] Request received.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from the backend!"))
	})
	log.Println("Backend server starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Backend server failed: %s", err)
	}
}

// --- Middleware ---
// 1. Define the middleware function.
// It takes an http.Handler as input and returns a new http.Handler.
// This allows middleware to be "chained" together.
func loggingMiddleware(next http.Handler) http.Handler {
	// http.HandlerFunc is an adapter that allows a regular function
	// to be used as an http.Handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Code here runs *before* the next handler.
		start := time.Now()
		log.Printf("[Middleware] Started %s %s", r.Method, r.URL.Path)

		// Call the next handler in the chain. This could be another
		// piece of middleware or the final proxy handler.
		next.ServeHTTP(w, r)

		// Code here runs *after* the next handler has finished.
		log.Printf("[Middleware] Completed in %v", time.Since(start))
	})
}

// --- Proxy Server ---
func proxyServer() {
	target, _ := url.Parse("http://localhost:8081")
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 2. Create the main proxy handler.
	// This is the core logic that the middleware will wrap.
	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Proxy] Forwarding request...")
		proxy.ServeHTTP(w, r)
	})

	// 3. Wrap the main handler with the middleware.
	// The loggingMiddleware now wraps our proxyHandler. When a request comes in,
	// it will go to the middleware first, which then calls the proxy handler.
	http.Handle("/", loggingMiddleware(proxyHandler))

	log.Println("Proxy server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Proxy server failed: %s", err)
	}
}

func main() {
	go backendServer()
	proxyServer()

	// To test:
	// 1. Run this file: `go run 5_proxy_with_middleware.go`
	// 2. Send a request to the proxy: `curl http://localhost:8080/`
	// 3. Observe the logs. You will see logs from the middleware both before
	//    and after the logs from the proxy and backend, showing the request
	//    flow through the chain.
}
