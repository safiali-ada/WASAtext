package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type loginRequest struct {
	Name string `json:"name"`
}

type loginResponse struct {
	Identifier string `json:"identifier"`
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,16}$`)

// doLogin handles user login/registration
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !usernameRegex.MatchString(req.Name) {
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	// Check if user exists
	user, err := rt.db.GetUserByUsername(req.Name)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var identifier string
	if user != nil {
		// User exists, return their identifier
		identifier = user.ID
	} else {
		// Create new user
		identifier = uuid.New().String()
		err = rt.db.CreateUser(identifier, req.Name)
		if err != nil {
			rt.baseLogger.WithError(err).Error("error creating user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loginResponse{Identifier: identifier})
}
