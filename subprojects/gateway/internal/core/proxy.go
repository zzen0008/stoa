package core

import (
	"bytes"
	"encoding/json"
	"io"
	"llm-gateway/internal/core/provider"
	"llm-gateway/internal/core/router"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

// modifyRequestBody rewrites the model name in the request body and returns the new body and the translated model name.
func modifyRequestBody(body []byte, providerName string) ([]byte, string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, "", err
	}

	var translatedModel string
	if model, ok := data["model"].(string); ok {
		// Strip the provider prefix
		translatedModel = strings.TrimPrefix(model, providerName+"/")
		data["model"] = translatedModel
	}

	newBody, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}

	return newBody, translatedModel, nil
}

// Proxy is the core engine that handles request routing and proxying.
type Proxy struct {
	providerManager *provider.Manager
	router          *router.Router
}

// NewProxy creates a new proxy.
func NewProxy(pm *provider.Manager, r *router.Router) *Proxy {
	return &Proxy{
		providerManager: pm,
		router:          r,
	}
}

// ServeHTTP handles the incoming request, selects a provider, and proxies the request.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read the body to determine the strategy.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	r.Body.Close() // We've read it, so close it.

	// Create a new reader for the router and the downstream request.
	bodyReader := bytes.NewReader(body)

	strategy, err := p.router.SelectStrategy(bodyReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Restore the body for the downstream request.
	r.Body = io.NopCloser(bytes.NewReader(body))

	// Get the original model name for logging
	var reqBody router.RequestBody
	json.Unmarshal(body, &reqBody) // Ignore error, router already validated it

	// Attempt to proxy the request using the fallback chain.
	for _, providerName := range strategy.Providers {
		providerConfig, ok := p.providerManager.GetConfig(providerName)
		if !ok {
			log.Printf("Provider '%s' not found or not enabled", providerName)
			continue
		}

		// Rewrite the request body for the downstream provider.
		modifiedBody, translatedModel, err := modifyRequestBody(body, providerName)
		if err != nil {
			http.Error(w, "Failed to modify request body", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(modifiedBody))
		r.ContentLength = int64(len(modifiedBody))

		targetURL, err := url.Parse(providerConfig.TargetURL)
		if err != nil {
			log.Printf("Error parsing target URL for provider %s: %v", providerName, err)
			continue
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.Host = targetURL.Host
				req.URL.Path = path.Join(targetURL.Path, r.URL.Path) // Join paths
				req.Header.Set("Authorization", "Bearer "+providerConfig.APIKey)
			},
		}

		// Log the detailed routing information
		log.Printf(
			`Routing request: model="%s" -> provider="%s" as model="%s"`,
			reqBody.Model,
			providerName,
			translatedModel,
		)

		proxy.ServeHTTP(w, r)
		// This is a simplification. A real implementation would need to inspect
		// the response status code before deciding to fall back.
		return // For now, we don't fall back.
	}

	http.Error(w, "All providers in the fallback chain failed", http.StatusServiceUnavailable)
}
