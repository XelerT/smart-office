package handlers

import (
	"encoding/json"
	"net/http"

	"smart-office/internal/platform/logger"
)

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(status)

	if data == nil {
		return
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Log.Errorf("Failed to encode JSON response: %v", err)
	}
}
