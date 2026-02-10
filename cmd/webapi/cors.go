package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies a CORS policy to the router.
// As per PDF spec: Allow all origins, Max-Age = 1 second
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Authorization",
			"Content-Type",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Allow all origins as specified in PDF
		handlers.AllowedOrigins([]string{"*"}),
		// Max-Age = 1 second as specified in PDF
		handlers.MaxAge(1),
	)(h)
}
