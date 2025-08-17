package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestChain ensures that middleware are executed in the correct order.
func TestChain(t *testing.T) {
	var result string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result += "A"
			next.ServeHTTP(w, r)
			result += "D"
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result += "B"
			next.ServeHTTP(w, r)
			result += "C"
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is the main handler, does nothing for this test.
	})

	chainedHandler := Chain(middleware1, middleware2)(handler)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	chainedHandler.ServeHTTP(w, req)

	expected := "ABCD"
	if result != expected {
		t.Errorf("Chain execution order was incorrect, got: %s, want: %s", result, expected)
	}
}

// TestLogging ensures the logging middleware writes the expected message.
func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	manager := NewManager(logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggingHandler := manager.Logging(handler)

	req := httptest.NewRequest("GET", "/testpath", nil)
	w := httptest.NewRecorder()

	loggingHandler.ServeHTTP(w, req)

	expectedLog := "received request: GET /testpath\n"
	if buf.String() != expectedLog {
		t.Errorf("Log message was incorrect, got: %q, want: %q", buf.String(), expectedLog)
	}
}
