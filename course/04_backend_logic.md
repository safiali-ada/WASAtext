# Module 4: Business Logic & Handlers

This module explores the heart of the application: `service/api/`. This is where the core logic resides.

## 4.1 Request Lifecycle
When a request hits `http.ListenAndServe`, it flows through:
1.  **FrontendHandler** (Static check) -> Pass to API Router.
2.  **CORS Middleware** (Allow Origin check).
3.  **Router (`httprouter`)**: Matches `/session` to `doLogin`.
4.  **Auth Middleware (`rt.wrap`)**: Validates Token.
5.  **Handler Function**: `doLogin`, `sendMessage`, etc.

## 4.2 The Router (`api-handler.go`)
This file maps URLs to Functions.
- **`func (rt *_router) Handler() http.Handler`**: Registers routes.
- **Example**: `rt.router.POST("/session", rt.wrap(rt.doLogin))`
    - `rt.wrap` is our custom middleware. It wraps `doLogin` with authentication logic.

## 4.3 Authentication Middleware (`auth.go`)
This is crucial. Instead of checking `Authorization` header in every single function, we do it once here.
- **Function**: `wrap(fn func(...))`
- **What it does**:
    1.  Reads `Authorization` header.
    2.  Validates format: `Bearer <token>`.
    3.  Parses Token (UUID).
    4.  Checks Database: `GetUserById(token)`.
    5.  If valid: Creates a `reqcontext.RequestContext`.
    6.  **Injects Context**: Pass the `UserID` to the next handler.
    7.  If invalid: Returns `401 Unauthorized`.

- **`reqcontext` Package**: A tiny helper struct to hold the `UserID`. This ensures we always know *who* is making the request in a type-safe way.

## 4.4 Login Logic (`login.go`)
- **Input**: JSON. We define a struct `User` to match the JSON input.
- **Logic**:
    - `json.NewDecoder(r.Body).Decode(&u)`: Parse JSON.
    - `db.GetUserByUsername(u.Username)`: Check if exists.
    - If exists -> Return `id`.
    - If not -> `db.CreateUser(u.Username)`.
- **Why Simplified?**: The requirements specified a username-only login. The "token" IS the User ID.

## 4.5 Messaging Logic (`messages.go`)
This handler serves `POST /conversations/:conversationId/messages`.
- **Input Handling**: Supports both JSON (text) and Multipart (images).
- **Security Check**: `db.IsUserInConversation(userId, conversationId)`.
    - *Why?* We must ensure Alice cannot post to Bob's private chat unless she is invited.
- **Creation**: Calls `db.CreateMessage`.
- **Response**: Returns the created message object (201 Created).

## 4.6 Conversation Logic (`conversations.go`)
- **`getMyConversations`**:
    - Calls `db.GetUserConversations(userId)`.
    - This is complex SQL (JOINs) encapsulated in a simple Go function call.
- **`getConversation`**:
    - Fetches messages.
    - **Mark as Read**: Critical feature. When Alice gets the conversation, we call `db.MarkConversationRead(conversationId, userId)`. This updates her `last_read_at` timestamp, which affects the checkmarks (Blue ticks) others see.
