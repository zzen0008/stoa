package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// This example demonstrates a fallback mechanism.
// The proxy will try to send the request to a primary backend. If that backend
// is unavailable (e.g., it's down or returns an error), the proxy will "fall back"
// and retry the request with a secondary backend.

// --- Backend Servers ---
func primaryBackend() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Primary Backend] Request received. Responding successfully.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from the PRIMARY backend!"))
	})
	log.Println("Primary backend server starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Primary backend failed: %s", err)
	}
}

func secondaryBackend() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Secondary Backend] Request received. Responding successfully.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from the FALLBACK backend!"))
	})
	log.Println("Secondary backend server starting on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatalf("Secondary backend failed: %s", err)
	}
}

// --- Proxy Server ---
func proxyServer() {
	primaryURL, _ := url.Parse("http://localhost:8081")
	secondaryURL, _ := url.Parse("http://localhost:8082")

	primaryProxy := httputil.NewSingleHostReverseProxy(primaryURL)
	secondaryProxy := httputil.NewSingleHostReverseProxy(secondaryURL)

	// 1. Define a custom ErrorHandler for the primary proxy.
	// This function is called whenever the proxy request to the backend fails.
	// This could be due to a network error (like connection refused) or the
	// backend responding with a 5xx status code.
	primaryProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[Proxy] Primary backend failed: %v", err)
		log.Println("[Proxy] Attempting to fall back to secondary backend...")

		// 2. In case of an error, serve the request using the secondary proxy.
		secondaryProxy.ServeHTTP(w, r)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Proxy] Forwarding request to primary backend...")
		primaryProxy.ServeHTTP(w, r)
	})

	log.Println("Proxy server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Proxy server failed: %s", err)
	}
}

func main() {
	// Start both backends.
	go primaryBackend()
	go secondaryBackend()

	// Give the servers a moment to start up.
	time.Sleep(100 * time.Millisecond)

	// Start the proxy.
	go proxyServer()

	log.Println("---")
	log.Println("TEST CASE 1: Primary backend is running.")
	log.Println("Sending request to proxy: curl http://localhost:8080/")
	log.Println("Expected: Response from PRIMARY backend.")
	log.Println("---")

	// Wait for user to see the first test case.
	time.Sleep(5 * time.Second)

	log.Println("---")
	log.Println("TEST CASE 2: Simulating primary backend failure.")
	log.Println("To simulate, you would stop the primaryBackend process.")
	log.Println("For this example, we can't stop the goroutine, but if you were to run")
	log.Println("the primary backend in a separate terminal and kill it, the next request")
	log.Println("would automatically go to the secondary backend.")
	log.Println("Sending request to proxy: curl http://localhost:8080/")
	log.Println("Expected (if primary is down): Response from FALLBACK backend.")
	log.Println("---")

	// Keep the main goroutine alive to allow for manual testing.
	select {}
}
