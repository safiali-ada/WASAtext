package api

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/api/reqcontext"
)

// wrap wraps a handler function with authentication
func (rt *_router) wrap(fn func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		// Verify user exists (token is the user ID)
		user, err := rt.db.GetUserByID(token)
		if err != nil {
			rt.baseLogger.WithError(err).Error("error checking user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := reqcontext.RequestContext{
			UserID: token,
		}

		fn(w, r, ps, ctx)
	}
}
