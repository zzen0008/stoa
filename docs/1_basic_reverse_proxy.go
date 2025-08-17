package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// This example demonstrates the simplest possible reverse proxy.
// The proxy forwards all incoming requests to a single downstream "backend" server.

// --- Backend Server ---
// A simple HTTP server that will act as our downstream service.
func backendServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Backend] Received request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from the backend server!"))
	})

	log.Println("Backend server starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Backend server failed: %s", err)
	}
}

// --- Proxy Server ---
// The reverse proxy that listens for client requests and forwards them.
func proxyServer() {
	// 1. Define the target URL of the backend server.
	target, err := url.Parse("http://localhost:8081")
	if err != nil {
		log.Fatalf("Failed to parse target URL: %s", err)
	}

	// 2. Create a new reverse proxy.
	// NewSingleHostReverseProxy returns a new ReverseProxy that routes
	// all incoming requests to the given target.
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 3. Create the main handler for the proxy server.
	// All requests to the proxy will be passed to the proxy.ServeHTTP method.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Proxy] Forwarding request: %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	log.Println("Proxy server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Proxy server failed: %s", err)
	}
}

func main() {
	// Run the backend server in a separate goroutine.
	go backendServer()

	// Run the proxy server in the main goroutine.
	proxyServer()

	// To test:
	// 1. Run this file: `go run 1_basic_reverse_proxy.go`
	// 2. Open a new terminal and send a request to the proxy:
	//    `curl http://localhost:8080/some/path`
	// 3. Observe the logs from both the proxy and the backend.
}
