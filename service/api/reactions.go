package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/api/reqcontext"
)

type commentRequest struct {
	Comment string `json:"comment"`
}

// commentMessage adds a comment/reaction to a message
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	// Get the message
	msg, err := rt.db.GetMessage(messageID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if msg == nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Check if user is a member of the conversation
	isMember, err := rt.db.IsConversationMember(msg.ConversationID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req commentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Comment) == 0 || len(req.Comment) > 50 {
		http.Error(w, "Comment must be 1-50 characters", http.StatusBadRequest)
		return
	}

	if err := rt.db.AddComment(messageID, ctx.UserID, req.Comment); err != nil {
		rt.baseLogger.WithError(err).Error("error adding comment")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// uncommentMessage removes a comment from a message
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	// Get the message
	msg, err := rt.db.GetMessage(messageID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if msg == nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Check if user is a member of the conversation
	isMember, err := rt.db.IsConversationMember(msg.ConversationID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := rt.db.RemoveComment(messageID, ctx.UserID); err != nil {
		rt.baseLogger.WithError(err).Error("error removing comment")
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
