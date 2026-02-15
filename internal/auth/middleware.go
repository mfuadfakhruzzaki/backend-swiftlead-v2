package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// Middleware creates an authentication middleware
func Middleware(secret string) func(http.Handler) http.Handler {
	       return func(next http.Handler) http.Handler {
		       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			       authHeader := r.Header.Get("Authorization")
			       if authHeader == "" {
				       // Fallback: check ?token= query param (required for WebSocket)
				       if tokenParam := r.URL.Query().Get("token"); tokenParam != "" {
					       authHeader = "Bearer " + tokenParam
				       } else {
					       response.Unauthorized(w, "Missing authorization header")
					       return
				       }
			       }

			       parts := strings.Split(authHeader, " ")
			       if len(parts) != 2 || parts[0] != "Bearer" {
				       response.Unauthorized(w, "Invalid authorization header format")
				       return
			       }

			       claims, err := ValidateToken(parts[1], secret)
			       if err != nil {
				       if err == ErrExpiredToken {
					       response.Unauthorized(w, "Token has expired")
				       } else {
					       response.Unauthorized(w, "Invalid token")
				       }
				       return
			       }

			       // Add claims to context
			       ctx := context.WithValue(r.Context(), UserContextKey, claims)
			       next.ServeHTTP(w, r.WithContext(ctx))
		       })
	       }
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims == nil {
				response.Unauthorized(w, "User not authenticated")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.Forbidden(w, "You don't have permission to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a shortcut for requiring admin role
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin)
}

// RequireAdminOrTechnician is a shortcut for requiring admin or technician role
func RequireAdminOrTechnician() func(http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin, models.RoleTechnician)
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
