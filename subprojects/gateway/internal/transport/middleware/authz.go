package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"llm-gateway/internal/config"

	"github.com/sirupsen/logrus"
)

// ModelRequest is used to unmarshal the model name from the request body.
type ModelRequest struct {
	Model string `json:"model"`
}

// Authorizer is a middleware that checks if a user, based on their groups,
// is authorized to use a specific model.
type Authorizer struct {
	log      *logrus.Logger
	providers []config.Provider
}

// NewAuthorizer creates a new Authorizer middleware.
func NewAuthorizer(log *logrus.Logger, providers []config.Provider) *Authorizer {
	return &Authorizer{
		log:      log,
		providers: providers,
	}
}

// Authorization is the middleware handler for authorization.
func (m *Manager) Authorization(authz *Authorizer) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Read the request body to get the model name.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				authz.log.Errorf("Failed to read request body: %v", err)
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			// Restore the body so it can be read again by the next handler.
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			var req ModelRequest
			if err := json.Unmarshal(body, &req); err != nil {
				authz.log.Errorf("Failed to unmarshal request body: %v", err)
				http.Error(w, "Invalid request format", http.StatusBadRequest)
				return
			}
			modelName := req.Model

			// 2. Get the user's groups from the context.
			userGroups, ok := r.Context().Value("user_groups").([]string)
			if !ok {
				// This should not happen if the auth middleware is configured correctly.
				authz.log.Error("User groups not found in context.")
				http.Error(w, "User groups not found", http.StatusInternalServerError)
				return
			}

			// 3. Find the model's allowed groups from the configuration.
			model, found := authz.findModel(modelName)
			if !found {
				authz.log.Warnf("Model '%s' not found in configuration.", modelName)
				http.Error(w, "Model not found", http.StatusNotFound)
				return
			}

			// If allowed_groups is empty, access is permitted for all authenticated users.
			if len(model.AllowedGroups) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// 4. Check for group membership.
			if !isAuthorized(userGroups, model.AllowedGroups) {
				authz.log.Warnf("User with groups %v is not authorized for model '%s'", userGroups, modelName)
				http.Error(w, "You are not authorized to use this model", http.StatusForbidden)
				return
			}

			authz.log.Infof("User with groups %v is authorized for model '%s'", userGroups, modelName)
			next.ServeHTTP(w, r)
		})
	}
}

// findModel searches for a model by name across all providers.
func (a *Authorizer) findModel(modelName string) (config.Model, bool) {
	for _, p := range a.providers {
		for _, m := range p.Models {
			if m.Name == modelName {
				return m, true
			}
		}
	}
	return config.Model{}, false
}

// isAuthorized checks if any of the user's groups are in the list of allowed groups.
func isAuthorized(userGroups, allowedGroups []string) bool {
	allowedMap := make(map[string]struct{}, len(allowedGroups))
	for _, g := range allowedGroups {
		allowedMap[g] = struct{}{}
	}

	for _, userGroup := range userGroups {
		if _, ok := allowedMap[userGroup]; ok {
			return true
		}
	}
	return false
}
