package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/api/reqcontext"
	"github.com/sapienzaapps/wasatext/service/database"
)

type messageResponse struct {
	ID             string                  `json:"id"`
	SenderID       string                  `json:"senderId"`
	SenderUsername string                  `json:"senderUsername"`
	Type           string                  `json:"type"`
	Content        string                  `json:"content"`
	Timestamp      string                  `json:"timestamp"`
	Checkmarks     int                     `json:"checkmarks"`
	ReplyTo        *messagePreviewResponse `json:"replyTo,omitempty"`
	Forwarded      bool                    `json:"forwarded"`
	Comments       []commentResponse       `json:"comments"`
}

type commentResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Comment  string `json:"comment"`
}

type sendMessageRequest struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	ReplyToID string `json:"replyToId,omitempty"`
}

type forwardMessageRequest struct {
	MessageID string `json:"messageId"`
}

// sendMessage sends a message to a conversation
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// Check membership
	isMember, err := rt.db.IsConversationMember(conversationID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	contentType := r.Header.Get("Content-Type")

	var msg database.Message
	msg.ID = uuid.New().String()
	msg.ConversationID = conversationID
	msg.SenderID = ctx.UserID

	if contentType == "application/json" {
		var req sendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Type != "text" && req.Type != "photo" {
			http.Error(w, "Invalid message type", http.StatusBadRequest)
			return
		}

		msg.Type = req.Type
		msg.Content = req.Content
		msg.ReplyToID = req.ReplyToID
	} else {
		// Handle multipart form for photo upload
		r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
		if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Error reading photo", http.StatusBadRequest)
			return
		}
		defer file.Close()

		photo, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading photo", http.StatusBadRequest)
			return
		}

		msg.Type = "photo"
		msg.Photo = photo
		msg.ReplyToID = r.FormValue("replyToId")
	}

	if err := rt.db.CreateMessage(&msg); err != nil {
		rt.baseLogger.WithError(err).Error("error creating message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get the created message info
	createdMsg, _ := rt.db.GetMessage(msg.ID)
	sender, _ := rt.db.GetUserByID(ctx.UserID)
	senderUsername := ""
	if sender != nil {
		senderUsername = sender.Username
	}

	response := messageResponse{
		ID:             msg.ID,
		SenderID:       msg.SenderID,
		SenderUsername: senderUsername,
		Type:           msg.Type,
		Timestamp:      createdMsg.CreatedAt,
		Checkmarks:     1,
		Forwarded:      false,
		Comments:       []commentResponse{},
	}

	if msg.Type == "text" {
		response.Content = msg.Content
	} else {
		response.Content = "/messages/" + msg.ID + "/photo"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("error encoding response")
	}
}

// forwardMessage forwards a message to a conversation
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// Check membership in target conversation
	isMember, err := rt.db.IsConversationMember(conversationID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req forwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the original message
	originalMsg, err := rt.db.GetMessage(req.MessageID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if originalMsg == nil {
		http.Error(w, "Original message not found", http.StatusNotFound)
		return
	}

	// Check if user can access the original message (is member of its conversation)
	isMemberOriginal, err := rt.db.IsConversationMember(originalMsg.ConversationID, ctx.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMemberOriginal {
		http.Error(w, "Original message not found", http.StatusNotFound)
		return
	}

	// Create forwarded message
	newMsg := database.Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		SenderID:       ctx.UserID,
		Content:        originalMsg.Content,
		Photo:          originalMsg.Photo,
		Type:           originalMsg.Type,
		Forwarded:      true,
	}

	if err := rt.db.CreateMessage(&newMsg); err != nil {
		rt.baseLogger.WithError(err).Error("error creating message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	createdMsg, _ := rt.db.GetMessage(newMsg.ID)
	sender, _ := rt.db.GetUserByID(ctx.UserID)
	senderUsername := ""
	if sender != nil {
		senderUsername = sender.Username
	}

	response := messageResponse{
		ID:             newMsg.ID,
		SenderID:       newMsg.SenderID,
		SenderUsername: senderUsername,
		Type:           newMsg.Type,
		Timestamp:      createdMsg.CreatedAt,
		Checkmarks:     1,
		Forwarded:      true,
		Comments:       []commentResponse{},
	}

	if newMsg.Type == "text" {
		response.Content = newMsg.Content
	} else {
		response.Content = "/messages/" + newMsg.ID + "/photo"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("error encoding response")
	}
}

// deleteMessage deletes a message
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// Only sender can delete
	if msg.SenderID != ctx.UserID {
		http.Error(w, "Forbidden - only sender can delete", http.StatusForbidden)
		return
	}

	if err := rt.db.DeleteMessage(messageID); err != nil {
		rt.baseLogger.WithError(err).Error("error deleting message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
