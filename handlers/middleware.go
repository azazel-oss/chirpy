package handlers

import (
	"chirpy/utils"
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// AuthMiddleware is a middleware function for JWT authentication
func (a *ApiConfig) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.parseJWT(r)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "the user is not authorized")
			return
		}

		// Store the claims in the context for later use
		ctx := context.WithValue(r.Context(), userClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserClaims(ctx context.Context) (*jwt.RegisteredClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*jwt.RegisteredClaims)
	return claims, ok
}

func (a *ApiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}
