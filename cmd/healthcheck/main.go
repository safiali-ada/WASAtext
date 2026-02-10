package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: healthcheck <url>")
		os.Exit(1)
	}

	url := os.Args[1]
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Healthcheck failed with status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("Healthcheck passed")
	os.Exit(0)
}
