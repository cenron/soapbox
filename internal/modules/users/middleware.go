package users

import (
	"net/http"
	"strings"

	"github.com/radni/soapbox/internal/core/httpkit"
	"github.com/radni/soapbox/internal/core/types"
)

var roleHierarchy = map[string]int{
	RoleUser:      0,
	RoleModerator: 1,
	RoleAdmin:     2,
}

func AuthRequired(tokens *TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpkit.Error(w, types.ErrUnauthorized())
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpkit.Error(w, types.ErrUnauthorized())
				return
			}

			claims, err := tokens.ValidateAccessToken(parts[1])
			if err != nil {
				httpkit.Error(w, types.ErrUnauthorized())
				return
			}

			userID, err := types.ParseID(claims.Subject)
			if err != nil {
				httpkit.Error(w, types.ErrUnauthorized())
				return
			}

			ctx := httpkit.WithUserID(r.Context(), userID)
			ctx = httpkit.WithRole(ctx, claims.Role)
			ctx = httpkit.WithVerified(ctx, claims.Verified)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthOptional(tokens *TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := tokens.ValidateAccessToken(parts[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := types.ParseID(claims.Subject)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := httpkit.WithUserID(r.Context(), userID)
			ctx = httpkit.WithRole(ctx, claims.Role)
			ctx = httpkit.WithVerified(ctx, claims.Verified)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleRequired(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := httpkit.RoleFrom(r.Context())
			if !ok {
				httpkit.Error(w, types.ErrForbidden())
				return
			}

			roleLevel, roleOK := roleHierarchy[role]
			minLevel, minOK := roleHierarchy[minRole]
			if !roleOK || !minOK || roleLevel < minLevel {
				httpkit.Error(w, types.ErrForbidden())
				return
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
