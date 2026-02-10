package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/api/reqcontext"
)

type createGroupRequest struct {
	Name string `json:"name"`
}

type addMemberRequest struct {
	UserID string `json:"userId"`
}

type setGroupNameRequest struct {
	Name string `json:"name"`
}

type groupResponse struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	PhotoURL *string        `json:"photoUrl,omitempty"`
	Members  []userResponse `json:"members"`
}

// createGroup creates a new group
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Name) == 0 || len(req.Name) > 50 {
		http.Error(w, "Group name must be 1-50 characters", http.StatusBadRequest)
		return
	}

	groupID := uuid.New().String()
	if err := rt.db.CreateGroupConversation(groupID, req.Name, ctx.UserID); err != nil {
		rt.baseLogger.WithError(err).Error("error creating group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get creator info
	creator, _ := rt.db.GetUserByID(ctx.UserID)
	members := []userResponse{}
	if creator != nil {
		members = append(members, userResponse{
			ID:       creator.ID,
			Username: creator.Username,
		})
	}

	response := groupResponse{
		ID:      groupID,
		Name:    req.Name,
		Members: members,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// addToGroup adds a user to the group
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// Check if group exists and is a group type
	conv, err := rt.db.GetConversation(groupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if conv == nil || conv.Type != "group" {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check if requester is a member
	isMember, err := rt.db.IsConversationMember(groupID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden - not a member", http.StatusForbidden)
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if target user exists
	targetUser, err := rt.db.GetUserByID(req.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := rt.db.AddGroupMember(groupID, req.UserID); err != nil {
		rt.baseLogger.WithError(err).Error("error adding member")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup removes a user from the group
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	userID := ps.ByName("userId")

	// Users can only remove themselves
	if userID != ctx.UserID {
		http.Error(w, "Forbidden - can only remove yourself", http.StatusForbidden)
		return
	}

	// Check if group exists
	conv, err := rt.db.GetConversation(groupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if conv == nil || conv.Type != "group" {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	if err := rt.db.RemoveGroupMember(groupID, userID); err != nil {
		rt.baseLogger.WithError(err).Error("error removing member")
		http.Error(w, "Not a member", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupName changes the group name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// Check if group exists
	conv, err := rt.db.GetConversation(groupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if conv == nil || conv.Type != "group" {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check membership
	isMember, err := rt.db.IsConversationMember(groupID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden - not a member", http.StatusForbidden)
		return
	}

	var req setGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Name) == 0 || len(req.Name) > 50 {
		http.Error(w, "Group name must be 1-50 characters", http.StatusBadRequest)
		return
	}

	if err := rt.db.UpdateGroupName(groupID, req.Name); err != nil {
		rt.baseLogger.WithError(err).Error("error updating group name")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupPhoto sets the group photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// Check if group exists
	conv, err := rt.db.GetConversation(groupID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if conv == nil || conv.Type != "group" {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check membership
	isMember, err := rt.db.IsConversationMember(groupID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden - not a member", http.StatusForbidden)
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

	if err := rt.db.UpdateGroupPhoto(groupID, photo); err != nil {
		rt.baseLogger.WithError(err).Error("error updating group photo")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
