package httpkit

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/radni/soapbox/internal/core/types"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("httpkit: failed to encode response", "error", err)
	}
}

func Error(w http.ResponseWriter, err error) {
	if appErr, ok := types.IsAppError(err); ok {
		JSON(w, appErr.Code, appErr)
		return
	}

	slog.Error("unhandled error", "error", err)
	JSON(w, http.StatusInternalServerError, map[string]string{
		"message": "internal server error",
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
