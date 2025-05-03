package main

import (
	"fmt"
	"net/http"

	"smart-office/internal/config"
	"smart-office/internal/platform/logger"
	"smart-office/internal/server"
)

func main() {
	logger.Init()
	logger.Log.Info("Starting application...")

	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatalf("Failed to load configuration: %v", err)
	}
	logger.Log.Infof("Configuration loaded successfully")

	// Data base setup
	// dbClient, err := database.Connect(cfg.MongoURI)

	mux := server.SetupRouter()

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Log.Infof("Starting server on %s", addr)

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		logger.Log.Fatalf("Failed to start server: %v", err)
	}
}
