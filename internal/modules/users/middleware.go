package users

import (
	"net/http"

	"github.com/radni/soapbox/internal/core/httpkit"
	"github.com/radni/soapbox/internal/core/types"
)

func AuthRequired(tokens *TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpkit.Error(w, types.ErrUnauthorized())
		})
	}
}

func RoleRequired(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpkit.Error(w, types.ErrForbidden())
		})
	}
}
