package core

import (
	"io"
	"llm-gateway/internal/config"
	"llm-gateway/internal/core/provider"
	"llm-gateway/internal/core/router"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockProviderServer creates a fake downstream LLM provider for testing.
func mockProviderServer(t *testing.T, expectedModel string, responseBody string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the model name was correctly rewritten.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("mock server failed to read request body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		// A simple check to see if the rewritten model name is in the body.
		// A real test might unmarshal the JSON for a more robust check.
		if !strings.Contains(string(body), `"model":"`+expectedModel+`"`) {
			t.Errorf("mock server received wrong model in body. got %s, want model %s", string(body), expectedModel)
			http.Error(w, "invalid model in body", http.StatusBadRequest)
			return
		}

		// Check for auth header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("mock server did not receive correct Authorization header")
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	})
	return httptest.NewServer(handler)
}

func TestProxyServeHTTPSuccess(t *testing.T) {
	// 1. Setup the mock server
	expectedModel := "test-model"
	responseBody := `{"id": "chatcmpl-123", "choices": [{"message": {"role": "assistant", "content": "Hello there!"}}]}`
	mockServer := mockProviderServer(t, expectedModel, responseBody)
	defer mockServer.Close()

	// 2. Configure the gateway to use the mock server
	cfg := config.Config{
		Providers: []config.Provider{
			{
				Name:      "mock-provider",
				Enabled:   true,
				TargetURL: mockServer.URL,
				APIKey:    "test-api-key",
				Timeout:   5 * time.Second,
			},
		},
		Strategies: []config.Strategy{
			{Name: "default", Providers: []string{"mock-provider"}},
		},
	}

	// 3. Initialize gateway components
	providerManager := provider.NewManager(cfg.Providers)
	coreRouter := router.NewRouter(cfg.Strategies)
	proxy := NewProxy(providerManager, coreRouter)

	// 4. Create the incoming request to the gateway
	requestBody := `{"model": "mock-provider/test-model", "messages": [{"role": "user", "content": "Hi"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(requestBody))
	rr := httptest.NewRecorder()

	// 5. Call the proxy handler
	proxy.ServeHTTP(rr, req)

	// 6. Assert the results
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Hello there!") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}
