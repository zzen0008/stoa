package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
)

// OIDCAuthenticator is a struct that holds the OIDC provider and verifier.
type OIDCAuthenticator struct {
	Provider *oidc.Provider
	Verifier *oidc.IDTokenVerifier
	Logger   *logrus.Logger
}

// NewOIDCAuthenticator creates a new OIDC authenticator.
func NewOIDCAuthenticator(ctx context.Context, logger *logrus.Logger, issuerURL, clientID string) (*OIDCAuthenticator, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCAuthenticator{
		Provider: provider,
		Verifier: verifier,
		Logger:   logger,
	}, nil
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

			idToken, err := auth.Verifier.Verify(r.Context(), rawToken)
			if err != nil {
				auth.Logger.Errorf("failed to verify token: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Token is valid, you can access claims from idToken if needed
			auth.Logger.Infof("successfully authenticated user: %s", idToken.Subject)

			next.ServeHTTP(w, r)
		})
	}
}
