package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/api/reqcontext"
)

type setUsernameRequest struct {
	Username string `json:"username"`
}

type userResponse struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	PhotoURL *string `json:"photoUrl,omitempty"`
}

// setMyUserName handles username change
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ps.ByName("userId")

	// Users can only change their own username
	if userID != ctx.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req setUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !usernameRegex.MatchString(req.Username) {
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	// Check if username is taken
	existingUser, err := rt.db.GetUserByUsername(req.Username)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking username")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil && existingUser.ID != userID {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	err = rt.db.UpdateUsername(userID, req.Username)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error updating username")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setMyPhoto handles profile photo upload
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ps.ByName("userId")

	// Users can only change their own photo
	if userID != ctx.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Limit to 5MB
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	photo, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading photo", http.StatusBadRequest)
		return
	}

	if len(photo) == 0 {
		http.Error(w, "Empty photo", http.StatusBadRequest)
		return
	}

	err = rt.db.UpdateUserPhoto(userID, photo)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error updating photo")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// searchUsers handles user search
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// An empty query lists every user, so the frontend can show all usernames
	query := r.URL.Query().Get("q")

	users, err := rt.db.SearchUsers(query)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error searching users")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]userResponse, len(users))
	for i, u := range users {
		response[i] = userResponse{
			ID:       u.ID,
			Username: u.Username,
			PhotoURL: photoURL(usersPathPrefix, u.ID, u.Photo),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("error encoding response")
	}
}

// getUserPhoto serves a user's profile photo without authentication.
func (rt *_router) getUserPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := ps.ByName("userId")

	photo, err := rt.db.GetUserPhoto(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting user photo")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(photo) == 0 {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(photo))
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := w.Write(photo); err != nil {
		rt.baseLogger.WithError(err).Error("error writing user photo")
	}
}
