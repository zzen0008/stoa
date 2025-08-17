package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// This example demonstrates how to modify a request before it's forwarded.
// The proxy will add a custom header to the request. This is the core technique
// for adding authentication tokens (like an API key) to requests sent to
// downstream providers.

// --- Backend Server ---
// A simple HTTP server that inspects and prints the headers it receives.
func backendServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Backend] Received request with headers:")
		// Print the custom header to prove it was added by the proxy.
		for name, values := range r.Header {
			if name == "X-Proxy-Signature" {
				log.Printf("[Backend]   -> %s: %s", name, values[0])
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Backend received the request."))
	})

	log.Println("Backend server starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Backend server failed: %s", err)
	}
}

// --- Proxy Server ---
func proxyServer() {
	target, err := url.Parse("http://localhost:8081")
	if err != nil {
		log.Fatalf("Failed to parse target URL: %s", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 1. Define a custom "Director" function.
	// The Director is a function that the ReverseProxy calls before forwarding
	// the request. It's the perfect place to modify the request.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Call the original director to set up the default proxy behavior.
		originalDirector(req)

		// 2. Modify the request object.
		// Here, we are adding a new header. In a real-world scenario, this is
		// where you would set the `Authorization` header with a provider's API key.
		log.Println("[Proxy] Adding custom header to the request")
		req.Header.Set("X-Proxy-Signature", "my-secret-proxy-token")

		// You can also modify the host, scheme, or path here.
		// For example, to route to a different provider:
		// req.URL.Scheme = "https"
		// req.URL.Host = "api.another-provider.com"
	}

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
	go backendServer()
	proxyServer()

	// To test:
	// 1. Run this file: `go run 2_proxy_with_request_modification.go`
	// 2. Send a request to the proxy: `curl http://localhost:8080/`
	// 3. Observe the logs. The backend should print the "X-Proxy-Signature" header.
}
