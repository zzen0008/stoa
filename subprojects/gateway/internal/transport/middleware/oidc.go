package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
)

// OIDCAuthenticator is a struct that holds the OIDC provider and verifier.
type OIDCAuthenticator struct {
	issuerURL string
	clientID  string
	logger    *logrus.Logger

	initOnce sync.Once
	verifier *oidc.IDTokenVerifier
	initErr  error
}

// NewOIDCAuthenticator creates a new OIDC authenticator.
func NewOIDCAuthenticator(logger *logrus.Logger, issuerURL, clientID string) *OIDCAuthenticator {
	return &OIDCAuthenticator{
		issuerURL: issuerURL,
		clientID:  clientID,
		logger:    logger,
	}
}

func (a *OIDCAuthenticator) getVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	a.initOnce.Do(func() {
		provider, err := oidc.NewProvider(ctx, a.issuerURL)
		if err != nil {
			a.initErr = err
			return
		}
		a.verifier = provider.Verifier(&oidc.Config{ClientID: a.clientID})
	})
	return a.verifier, a.initErr
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

			// Token is valid, you can access claims from idToken if needed
			auth.logger.Infof("successfully authenticated user: %s", idToken.Subject)

			next.ServeHTTP(w, r)
		})
	}
}
