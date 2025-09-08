package middleware

import (
	"context"
	"llm-gateway/internal/config"
	"llm-gateway/internal/ratelimit"
	"net/http"
	"sort"

	"github.com/sirupsen/logrus"
)

// RateLimiter is the middleware handler for rate limiting.
func (m *Manager) RateLimiter(store ratelimit.RateLimiterStore, cfg config.RateLimit) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract user groups from request context (set by OIDC middleware).
			groups, ok := r.Context().Value("user_groups").([]string)
			if !ok {
				// If no groups are found, apply the default rate limit based on IP
				// For simplicity, we will use a generic "unauthenticated" key.
				// A better approach for production would be to use the IP address.
				key := "ratelimit:unauthenticated"
				handleRateLimit(w, r, next, store, key, cfg.Default, m.Logger)
				return
			}

			// 2. Determine the most restrictive rate limit for the user's groups.
			// Sort groups to ensure consistent behavior if a user is in multiple groups with the same limit.
			sort.Strings(groups)

			// Start with the default limit
			finalLimit := cfg.Default
			// Assign a default group name in case no specific group matches
			finalLimit.Name = "default"

			// Find the most restrictive limit among the user's groups
			for _, group := range groups {
				if groupLimit, exists := cfg.Groups[group]; exists {
					// Lower requests per window is more restrictive
					if groupLimit.Requests < finalLimit.Requests {
						finalLimit = groupLimit
						finalLimit.Name = group // Set the name to the matched group
					}
				}
			}

			// 3. The key for the rate limiter should include both the group name and user ID.
			// This ensures each user has their own rate limit within the group.
			userID, ok := r.Context().Value("user_id").(string)
			if !ok {
				// Fallback to group-only key if user ID is not available
				key := "ratelimit:" + finalLimit.Name
				// 4. Check with the store.
				handleRateLimit(w, r, next, store, key, finalLimit, m.Logger)
				return
			}

			key := "ratelimit:" + finalLimit.Name + ":" + userID

			// 4. Check with the store.
			handleRateLimit(w, r, next, store, key, finalLimit, m.Logger)
		})
	}
}

func handleRateLimit(w http.ResponseWriter, r *http.Request, next http.Handler, store ratelimit.RateLimiterStore, key string, limit config.RateLimitConfig, logger *logrus.Logger) {
	allowed, err := store.Allow(r.Context(), key, limit.Requests, limit.Window)
	if err != nil {
		// Log the error and fail open (allow the request) to avoid blocking users due to a backend issue.
		logger.Errorf("Rate limiter error for key %s: %v", key, err)
		next.ServeHTTP(w, r)
		return
	}

	if !allowed {
		logger.Warnf("Rate limit exceeded for key %s", key)
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	next.ServeHTTP(w, r)
}

// Helper function to get user groups from context (improves on the original direct casting)
func GetUserGroups(ctx context.Context) ([]string, bool) {
	groups, ok := ctx.Value("user_groups").([]string)
	return groups, ok
}
