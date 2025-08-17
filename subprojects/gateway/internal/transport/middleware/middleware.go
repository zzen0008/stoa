package middleware

import (
	"log"
	"net/http"
)

// Middleware is a function that takes an http.Handler and returns a new http.Handler.
// It's used for pre-forwarding logic at the transport layer.
type Middleware func(http.Handler) http.Handler

// Chain creates a new middleware that applies a series of middlewares in order.
func Chain(middlewares ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}
		return h
	}
}

// Manager holds dependencies for transport-layer middleware.
type Manager struct {
	Logger *log.Logger
}

// NewManager creates a new Manager.
func NewManager(logger *log.Logger) *Manager {
	return &Manager{
		Logger: logger,
	}
}

// Logging logs the incoming HTTP request.
func (m *Manager) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Logger.Printf("received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
