package main

import (
	"os"
	"time"
)

// WebAPIConfiguration is the configuration for the web API server
type WebAPIConfiguration struct {
	Web struct {
		APIHost         string
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
	}
	DB struct {
		Filename string
	}
	Debug bool
}

func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Load from environment or use defaults matching Dockerfile
	cfg.Web.APIHost = os.Getenv("WASATEXT_WEB_APIHOST")
	if cfg.Web.APIHost == "" {
		cfg.Web.APIHost = ":3000"
	}

	cfg.DB.Filename = os.Getenv("WASATEXT_DB_FILENAME")
	if cfg.DB.Filename == "" {
		cfg.DB.Filename = "./wasatext.db"
	}

	// Hardcoded timeouts for simplicity
	cfg.Web.ReadTimeout = 5 * time.Second
	cfg.Web.WriteTimeout = 5 * time.Second
	cfg.Web.ShutdownTimeout = 5 * time.Second

	if os.Getenv("WASATEXT_DEBUG") == "true" {
		cfg.Debug = true
	} else {
		cfg.Debug = false
	}

	return cfg, nil
}
