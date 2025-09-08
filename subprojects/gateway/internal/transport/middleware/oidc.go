package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
)

// OIDCAuthenticator is a struct that holds the OIDC provider and verifier.
type OIDCAuthenticator struct {
	issuerURL    string
	clientID     string
	logger       *logrus.Logger
	cacheTTL     time.Duration
	verifier     *oidc.IDTokenVerifier
	verifierMu   sync.RWMutex
	lastVerified time.Time
}

// NewOIDCAuthenticator creates a new OIDC authenticator.
func NewOIDCAuthenticator(logger *logrus.Logger, issuerURL, clientID string, cacheTTL time.Duration) *OIDCAuthenticator {
	return &OIDCAuthenticator{
		issuerURL: issuerURL,
		clientID:  clientID,
		logger:    logger,
		cacheTTL:  cacheTTL,
	}
}

func (a *OIDCAuthenticator) getVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	a.verifierMu.RLock()
	if a.verifier != nil && time.Since(a.lastVerified) < a.cacheTTL {
		a.verifierMu.RUnlock()
		return a.verifier, nil
	}
	a.verifierMu.RUnlock()

	a.verifierMu.Lock()
	defer a.verifierMu.Unlock()

	// Double check if another goroutine has already refreshed the verifier
	if a.verifier != nil && time.Since(a.lastVerified) < a.cacheTTL {
		return a.verifier, nil
	}

	provider, err := oidc.NewProvider(ctx, a.issuerURL)
	if err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: a.clientID})
	a.verifier = verifier
	a.lastVerified = time.Now()

	return a.verifier, nil
}

// Authentication is the middleware handler for OIDC authentication.
func (m *Manager) Authentication(auth *OIDCAuthenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				http.Error(w, "Authorization header must be in the format 'Bearer {token}'", http.StatusUnauthorized)
				return
			}
			rawToken := tokenParts[1]

			verifier, err := auth.getVerifier(r.Context())
			if err != nil {
				auth.logger.Errorf("failed to initialize oidc verifier: %v", err)
				http.Error(w, "OIDC provider is unavailable", http.StatusServiceUnavailable)
				return
			}

			idToken, err := verifier.Verify(r.Context(), rawToken)
			if err != nil {
				auth.logger.Errorf("failed to verify token: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// The token signature is now verified. We will manually decode the payload
			// to robustly extract claims, as the go-oidc library's Claims() method
			// can be unreliable for custom claims in access tokens.
			payload, err := decodeJWTPayload(rawToken)
			if err != nil {
				auth.logger.Errorf("failed to decode token payload: %v", err)
				http.Error(w, "Failed to parse token claims", http.StatusUnauthorized)
				return
			}

			var claims struct {
				Groups []string `json:"groups"`
			}
			if err := json.Unmarshal(payload, &claims); err != nil {
				auth.logger.Errorf("failed to unmarshal custom claims: %v", err)
				http.Error(w, "Failed to parse token claims", http.StatusUnauthorized)
				return
			}

			// Store the user's groups and ID in the request context for downstream middleware.
			ctxWithGroups := context.WithValue(r.Context(), "user_groups", claims.Groups)
			ctxWithUserID := context.WithValue(ctxWithGroups, "user_id", idToken.Subject)
			r = r.WithContext(ctxWithUserID)

			// Token is valid, you can access claims from idToken if needed
			auth.logger.Infof("successfully authenticated user: %s", idToken.Subject)

			next.ServeHTTP(w, r)
		})
	}
}

// decodeJWTPayload decodes the payload part of a JWT string.
func decodeJWTPayload(token string) ([]byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, http.ErrAbortHandler
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	return payload, nil
}
