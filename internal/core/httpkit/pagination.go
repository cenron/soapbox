package httpkit

import (
	"net/http"

	"github.com/radni/soapbox/internal/core/types"
)

func CursorResponse[T any](w http.ResponseWriter, page types.CursorPage[T]) {
	JSON(w, http.StatusOK, page)
}
