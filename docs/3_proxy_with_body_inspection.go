package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// This example demonstrates how to inspect the request body before proxying.
// This is essential for our intelligent routing, where we need to read the
// "model" field from the JSON body to decide which provider to route to.

// The key challenge is that an http.Request.Body is an io.ReadCloser, which
// can only be read once. After we read it to inspect it, the body is empty.
// The solution is to read the body into a buffer, and then create a *new*
// io.ReadCloser from that buffer to attach back to the request.

// --- Backend Server ---
// A simple server that echoes the request body it receives.
func backendServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[Backend] Error reading body: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("[Backend] Received request with body: %s", string(body))
		w.WriteHeader(http.StatusOK)
		w.Write(body) // Echo the body back
	})

	log.Println("Backend server starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Backend server failed: %s", err)
	}
}

// A simple struct to represent the incoming JSON request.
type ChatRequest struct {
	Model string `json:"model"`
}

// --- Proxy Server ---
func proxyServer() {
	target, err := url.Parse("http://localhost:8081")
	if err != nil {
		log.Fatalf("Failed to parse target URL: %s", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	// The main handler that will inspect the body.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Proxy] Received request for inspection: %s %s", r.Method, r.URL.Path)

		// 1. Read the body into a byte slice.
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[Proxy] Error reading body: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// After this, r.Body is empty.

		// 2. Create a new reader from the byte slice and put it back on the request.
		// This makes the body readable again for the downstream proxy.
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 3. Now we can safely use the `bodyBytes` for our logic.
		// We'll unmarshal it to inspect the "model" field.
		var chatReq ChatRequest
		if err := json.Unmarshal(bodyBytes, &chatReq); err != nil {
			log.Printf("[Proxy] Error unmarshaling JSON: %s", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 4. Perform routing logic based on the model.
		// (In this toy example, we just log it).
		log.Printf("[Proxy] Routing decision: model is '%s'", chatReq.Model)
		if chatReq.Model == "openai/gpt-4o" {
			log.Printf("[Proxy] -> This would be routed to the OpenAI provider.")
		} else {
			log.Printf("[Proxy] -> This would be routed to a different provider.")
		}

		// 5. Forward the request. The proxy can now read the restored body.
		log.Println("[Proxy] Forwarding request to backend...")
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
	// 1. Run this file: `go run 3_proxy_with_body_inspection.go`
	// 2. Send a POST request with a JSON body:
	//    `curl -X POST -d '{"model": "openai/gpt-4o"}' http://localhost:8080/`
	// 3. Observe the logs. The proxy should log the model name, and the backend
	//    should log the full body it received, proving it was correctly forwarded.
}
