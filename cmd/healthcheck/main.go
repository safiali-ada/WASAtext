package main

import (
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	if err := run(logger); err != nil {
		logger.WithError(err).Error("healthcheck failed")
		os.Exit(1)
	}
}

func run(logger *logrus.Logger) error {
	if len(os.Args) < 2 {
		logger.Error("Usage: healthcheck <url>")
		return nil
	}

	url := os.Args[1]
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("healthcheck failed with status: %d", resp.StatusCode)
		return nil
	}

	logger.Info("Healthcheck passed")
	return nil
}
