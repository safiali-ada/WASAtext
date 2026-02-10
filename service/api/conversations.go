package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/api/reqcontext"
)

type conversationPreviewResponse struct {
	ID            string                  `json:"id"`
	Type          string                  `json:"type"`
	Name          string                  `json:"name"`
	PhotoURL      *string                 `json:"photoUrl,omitempty"`
	LatestMessage *messagePreviewResponse `json:"latestMessage,omitempty"`
}

type messagePreviewResponse struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	SenderID  string `json:"senderId"`
}

type conversationResponse struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	PhotoURL *string           `json:"photoUrl,omitempty"`
	Members  []userResponse    `json:"members"`
	Messages []messageResponse `json:"messages"`
}

type startConversationRequest struct {
	UserID string `json:"userId"`
}

// getMyConversations returns all conversations for the user
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ps.ByName("userId")

	if userID != ctx.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	previews, err := rt.db.GetUserConversations(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting conversations")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]conversationPreviewResponse, len(previews))
	for i, p := range previews {
		response[i] = conversationPreviewResponse{
			ID:   p.ID,
			Type: p.Type,
			Name: p.Name,
		}
		if len(p.Photo) > 0 {
			var photoURL string
			if p.Type == "group" {
				photoURL = "/groups/" + p.ID + "/photo"
			} else {
				photoURL = "/users/" + p.ID + "/photo"
			}
			response[i].PhotoURL = &photoURL
		}
		if p.LatestMessage != nil {
			response[i].LatestMessage = &messagePreviewResponse{
				Content:   p.LatestMessage.Content,
				Timestamp: p.LatestMessage.Timestamp,
				SenderID:  p.LatestMessage.SenderID,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getConversation returns a conversation with messages
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// Check if user is a member
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

	conv, err := rt.db.GetConversation(conversationID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if conv == nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Mark as read
	rt.db.MarkConversationRead(conversationID, ctx.UserID)

	// Get members
	members, err := rt.db.GetGroupMembers(conversationID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting members")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	memberResponses := make([]userResponse, len(members))
	for i, m := range members {
		memberResponses[i] = userResponse{
			ID:       m.ID,
			Username: m.Username,
		}
		if len(m.Photo) > 0 {
			photoURL := "/users/" + m.ID + "/photo"
			memberResponses[i].PhotoURL = &photoURL
		}
	}

	// Get messages
	messages, err := rt.db.GetConversationMessages(conversationID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting messages")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build message responses with checkmarks and comments
	messageResponses := make([]messageResponse, len(messages))
	for i, msg := range messages {
		checkmarks, _ := rt.db.GetMessageCheckmarks(msg.ID)
		comments, _ := rt.db.GetMessageComments(msg.ID)

		// Get sender username
		sender, _ := rt.db.GetUserByID(msg.SenderID)
		senderUsername := ""
		if sender != nil {
			senderUsername = sender.Username
		}

		messageResponses[i] = messageResponse{
			ID:             msg.ID,
			SenderID:       msg.SenderID,
			SenderUsername: senderUsername,
			Type:           msg.Type,
			Timestamp:      msg.CreatedAt,
			Checkmarks:     checkmarks,
			Forwarded:      msg.Forwarded,
			Comments:       make([]commentResponse, len(comments)),
		}

		if msg.Type == "text" {
			messageResponses[i].Content = msg.Content
		} else {
			messageResponses[i].Content = "/messages/" + msg.ID + "/photo"
		}

		for j, c := range comments {
			messageResponses[i].Comments[j] = commentResponse{
				UserID:   c.UserID,
				Username: c.Username,
				Comment:  c.Comment,
			}
		}

		if msg.ReplyToID != "" {
			replyTo, _ := rt.db.GetMessage(msg.ReplyToID)
			if replyTo != nil {
				replySender, _ := rt.db.GetUserByID(replyTo.SenderID)
				replySenderID := replyTo.SenderID
				if replySender != nil {
					messageResponses[i].ReplyTo = &messagePreviewResponse{
						Content:   replyTo.Content,
						Timestamp: replyTo.CreatedAt,
						SenderID:  replySenderID,
					}
				}
			}
		}
	}

	// Determine conversation name
	name := conv.GroupName
	if conv.Type == "private" {
		for _, m := range members {
			if m.ID != ctx.UserID {
				name = m.Username
				break
			}
		}
	}

	response := conversationResponse{
		ID:       conv.ID,
		Type:     conv.Type,
		Name:     name,
		Members:  memberResponses,
		Messages: messageResponses,
	}

	if len(conv.Photo) > 0 {
		photoURL := "/groups/" + conv.ID + "/photo"
		response.PhotoURL = &photoURL
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// startConversation creates or returns a private conversation
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req startConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == ctx.UserID {
		http.Error(w, "Cannot start conversation with yourself", http.StatusBadRequest)
		return
	}

	// Check if target user exists
	targetUser, err := rt.db.GetUserByID(req.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check for existing conversation
	existing, err := rt.db.GetPrivateConversation(ctx.UserID, req.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking existing conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existing != nil {
		response := conversationPreviewResponse{
			ID:   existing.ID,
			Type: "private",
			Name: targetUser.Username,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create new conversation
	convID := uuid.New().String()
	err = rt.db.CreatePrivateConversation(convID, ctx.UserID, req.UserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error creating conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := conversationPreviewResponse{
		ID:   convID,
		Type: "private",
		Name: targetUser.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
