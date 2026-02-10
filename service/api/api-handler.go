package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Handler returns an instance of httprouter.Router that handles APIs
func (rt *_router) Handler() http.Handler {
	// Login - no auth required
	rt.router.POST("/session", rt.doLogin)

	// User routes
	rt.router.PUT("/users/:userId/username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/:userId/photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/users", rt.wrap(rt.searchUsers))

	// Conversation routes
	rt.router.GET("/users/:userId/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversation))

	// Message routes
	rt.router.POST("/conversations/:conversationId/messages", rt.wrap(rt.sendMessage))
	rt.router.POST("/conversations/:conversationId/messages/forward", rt.wrap(rt.forwardMessage))
	rt.router.DELETE("/messages/:messageId", rt.wrap(rt.deleteMessage))

	// Reaction routes
	rt.router.PUT("/messages/:messageId/comment", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/messages/:messageId/comment", rt.wrap(rt.uncommentMessage))

	// Group routes
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.POST("/groups/:groupId/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId/members/:userId", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))

	// Liveness check
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}

func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusNoContent)
}
