package server

import (
	"fmt"
	"net/http"
	"smart-office/internal/platform/logger"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// --- Health Check Route ---
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"message": "pong"}`)
		logger.Log.Info("Handled /ping request")
	})

	// Place to register handlers

	return mux
}
