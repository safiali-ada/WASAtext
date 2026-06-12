# WASAText Oral Exam Guide

This guide explains the WASAText project from zero. It is written so you can speak about the project clearly even if the professor asks broad architecture questions, code-level questions, or "why did you do it this way?" questions.

## 1. One-Minute Project Explanation

WASAText is a full-stack messaging web application inspired by WhatsApp. It lets users log in with a username, search other users, start private chats, create groups, send text or photo messages, reply to messages, forward messages, delete their own messages, add simple reactions/comments, update profile photos, and manage group names or members.

The project has three main parts:

- A Go backend that exposes a REST API.
- A Vue 3 frontend that runs in the browser as a single page application.
- A SQLite database that stores users, conversations, members, messages, photos, and comments.

The application can run locally or inside Docker. In the main Docker setup, the Vue frontend is built into static files and served by the same Go server that handles the API.

Good oral answer:

> WASAText is a layered monolithic web application. The browser runs a Vue single page app. The Vue app calls a Go REST API using Axios. The Go API validates requests, checks authorization, and calls a database layer. The database layer stores everything in SQLite. The project is monolithic because one backend process serves both the API and the frontend, but internally it is separated into API and database layers.

## 2. Project Structure

Important folders:

```text
cmd/
  webapi/
    main.go                 backend entry point
    load-configuration.go   environment-based configuration
    frontend.go             serves Vue build files and handles SPA fallback
    cors.go                 CORS middleware
  healthcheck/
    main.go                 small HTTP healthcheck utility

service/
  api/
    api.go                  router construction and dependency injection
    api-handler.go          route registration
    auth.go                 Bearer token authentication wrapper
    login.go                login/register endpoint
    user.go                 username, profile photo, user search
    conversations.go        private chat and conversation retrieval
    messages.go             send, forward, delete, photo retrieval
    groups.go               group creation and management
    reactions.go            message comments/reactions
    constants.go            API constants
    reqcontext/             authenticated request context

  database/
    database.go             database interface, models, table creation
    users.go                user SQL operations
    conversations.go        conversation and group SQL operations
    messages.go             message and comment SQL operations

webui/
  src/
    main.js                 Vue app bootstrap
    router/index.js         frontend routes and auth guard
    services/axios.js       Axios instance and auth interceptor
    views/                  main screens
    assets/main.css         styling

doc/
  api.yaml                  OpenAPI API contract

Dockerfile                  full production image, frontend + backend
docker-compose.yml          easy local Docker run with persistent volume
```

## 3. Technologies Used

### Go Backend

Go is used for the backend because it is good for HTTP services, compiles to a single binary, has a strong standard library, and works well with simple deployment. The project uses:

- `net/http` for HTTP serving.
- `httprouter` for path-based routing.
- `logrus` for logging.
- `go-sqlite3` for SQLite access.
- `gorilla/handlers` only for CORS middleware.
- `google/uuid` to generate IDs.

Important correction: the README says "Gorilla Mux", but the actual router used by the API is `github.com/julienschmidt/httprouter`. Gorilla is only used for CORS helpers.

### Vue Frontend

Vue 3 is used for the frontend. It creates a single page application, meaning the browser loads one HTML page and Vue changes the screen without full page reloads. The frontend uses:

- Vue components for screens.
- Vue Router for navigation.
- Axios for API calls.
- Local storage for the simple login token.
- Vite for development and build.

### SQLite Database

SQLite is used because it is embedded, simple, portable, and appropriate for a course project. It stores data in a file, not in a separate database server.

Good oral answer:

> SQLite was chosen because this is a portable academic project. It avoids needing a separate PostgreSQL or MySQL service. For a small messaging app demo, SQLite is enough and keeps deployment simple.

## 4. Architecture

The architecture is a layered monolith.

```text
Browser
  |
  | Vue + Axios HTTP requests
  v
Go HTTP Server
  |
  | route + auth + validation
  v
API Handlers
  |
  | calls interface methods
  v
Database Layer
  |
  | SQL queries
  v
SQLite file
```

It is called a monolith because the backend is one Go process. It is called layered because responsibilities are separated:

- `cmd/webapi` starts and configures the server.
- `service/api` understands HTTP requests and responses.
- `service/database` understands SQL and persistence.
- `webui` is the browser application.

Why this is good:

- Easier to deploy than microservices.
- Easier to debug.
- No network calls between internal services.
- Still organized enough that code is not mixed together.

Professor may ask: "Why not microservices?"

Answer:

> Microservices would be unnecessary for this project size. We would need service discovery, inter-service communication, more Docker services, and distributed debugging. A layered monolith gives us clean separation without that operational complexity.

## 5. Backend Startup Flow

The backend starts in `cmd/webapi/main.go`.

Step by step:

1. `main()` creates a logger and calls `run()`.
2. `run()` loads configuration from environment variables.
3. It opens the SQLite database using `sql.Open("sqlite3", cfg.DB.Filename)`.
4. It passes the SQL connection to `database.New(dbconn)`.
5. `database.New` creates tables if they do not exist.
6. It creates the API router using `api.New`.
7. It calls `apirouter.Handler()` to register all HTTP routes.
8. It wraps the API with `FrontendHandler`, so the same server can serve Vue files.
9. It wraps everything with CORS middleware.
10. It starts `http.Server` on port `:3000` by default.
11. It listens for shutdown signals like Ctrl+C and tries graceful shutdown.

Good oral answer:

> The entry point is `cmd/webapi/main.go`. It loads config, opens SQLite, initializes the database schema, builds the API router with dependency injection, wraps it with frontend static file serving and CORS, then starts an HTTP server. It also handles graceful shutdown using OS signals.

## 6. Configuration

Configuration is loaded in `cmd/webapi/load-configuration.go`.

Important environment variables:

- `WASATEXT_WEB_APIHOST`: server address, default `:3000`.
- `WASATEXT_DB_FILENAME`: SQLite file path, default `./wasatext.db`.
- `WASATEXT_DEBUG`: if `"true"`, log level is debug.

Timeouts are hardcoded:

- Read timeout: 5 seconds.
- Write timeout: 30 seconds.
- Shutdown timeout: 5 seconds.

Why environment variables?

> Environment variables allow the same compiled binary to run in different environments. Locally it can use `./wasatext.db`; in Docker it uses `/app/data/wasatext.db`.

## 7. API Routing

Routes are registered in `service/api/api-handler.go`.

Main routes:

```text
POST   /session                                      login or create user
PUT    /users/:userId/username                       change username
PUT    /users/:userId/photo                          upload profile photo
GET    /users/:userId/photo                          download profile photo
GET    /users?q=...                                  search users

GET    /users/:userId/conversations                  list my conversations
POST   /conversations                                start private conversation
GET    /conversations/:conversationId                get messages and members

POST   /conversations/:conversationId/messages       send message
POST   /conversations/:conversationId/messages/forward forward message
DELETE /messages/:messageId                          delete own message
GET    /messages/:messageId/photo                    download message photo

PUT    /messages/:messageId/comment                  add/update reaction/comment
DELETE /messages/:messageId/comment                  remove reaction/comment

POST   /groups                                       create group
POST   /groups/:groupId/members                      add member
DELETE /groups/:groupId/members/:userId              leave group
PUT    /groups/:groupId/name                         rename group
PUT    /groups/:groupId/photo                        upload group photo
GET    /groups/:groupId/photo                        download group photo

GET    /liveness                                     liveness endpoint
```

All routes except `/session` and `/liveness` are wrapped with authentication.

## 8. Authentication

Authentication is implemented in `service/api/auth.go`.

The project uses simplified Bearer authentication:

```text
Authorization: Bearer <user-id>
```

The token is the user's UUID. The middleware:

1. Reads the `Authorization` header.
2. Checks it has the format `Bearer <token>`.
3. Uses the token as a user ID.
4. Calls `GetUserByID(token)`.
5. If the user exists, it creates a request context with `UserID`.
6. If not, it returns `401 Unauthorized`.

Important:

- There are no passwords.
- There are no signed JWTs.
- This is acceptable for a simplified course project but not secure enough for a real production app.

Good oral answer:

> The authentication is intentionally simple. During login, the backend returns the user UUID. The frontend stores it in local storage and sends it as a Bearer token. The backend validates the token by checking whether that user ID exists in the database.

Professor may ask: "Is this secure?"

Answer:

> Not for production. Anyone who knows a user ID could impersonate that user. A real app would use password authentication, session cookies or signed JWTs, HTTPS, token expiration, refresh tokens, and probably server-side session invalidation.

## 9. Login Flow

Backend file: `service/api/login.go`  
Frontend file: `webui/src/views/LoginView.vue`

Frontend:

1. User enters a username.
2. Vue calls `POST /session` with JSON:

```json
{ "name": "alice" }
```

Backend:

1. Decodes the JSON body.
2. Validates the username using regex:

```text
^[a-zA-Z0-9_]{3,16}$
```

3. Checks if the username already exists.
4. If yes, returns the existing user ID.
5. If no, generates a new UUID and inserts the user.
6. Returns:

```json
{ "identifier": "some-uuid" }
```

Frontend then stores:

- `wasatext_token`
- `wasatext_user_id`
- `wasatext_username`

Professor may ask: "Why does login return 201 even if the user already exists?"

Possible answer:

> The endpoint represents session creation, so it returns `201 Created` for a successful login/session creation. For strict REST semantics, returning `200 OK` for existing users and `201 Created` for new users could also be reasonable.

## 10. Database Schema

Tables are created in `service/database/database.go`.

### users

```sql
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  photo BLOB
);
```

Stores users.

- `id`: UUID string.
- `username`: unique login/display name.
- `photo`: binary profile image.

### conversations

```sql
CREATE TABLE conversations (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  group_name TEXT,
  photo BLOB
);
```

Stores both private and group conversations.

- `type`: `"private"` or `"group"`.
- `group_name`: only used for groups.
- `photo`: group photo.

### conversation_members

```sql
CREATE TABLE conversation_members (
  conversation_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  last_read_at DATETIME,
  PRIMARY KEY (conversation_id, user_id),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

This is a junction table. It models the many-to-many relationship between users and conversations:

- One user can belong to many conversations.
- One conversation can have many users.

The composite primary key prevents the same user from being inserted twice in the same conversation.

`last_read_at` is used for read receipts/checkmarks.

### messages

```sql
CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL,
  sender_id TEXT NOT NULL,
  content TEXT,
  photo BLOB,
  type TEXT NOT NULL,
  reply_to_id TEXT,
  forwarded INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id),
  FOREIGN KEY (sender_id) REFERENCES users(id)
);
```

Stores messages.

- `content`: text message content.
- `photo`: binary image for photo messages.
- `type`: `"text"` or `"photo"`.
- `reply_to_id`: points to another message.
- `forwarded`: 0 or 1 boolean flag.
- `created_at`: timestamp from SQLite.

### message_comments

```sql
CREATE TABLE message_comments (
  message_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  comment TEXT NOT NULL,
  PRIMARY KEY (message_id, user_id),
  FOREIGN KEY (message_id) REFERENCES messages(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

Stores message comments/reactions. The primary key means each user can have only one comment/reaction per message. The code uses `INSERT OR REPLACE`, so adding another reaction from the same user replaces the old one.

## 11. Database Layer and Interface

The file `service/database/database.go` defines `AppDatabase`, an interface with all database operations.

Why use an interface?

> The API layer does not need to know SQL details. It calls methods like `CreateMessage` or `GetUserConversations`. This separates HTTP logic from persistence logic and makes testing easier because a mock database could be injected.

The actual implementation is `appdbimpl`, which contains:

```go
type appdbimpl struct {
  c *sql.DB
}
```

So the implementation wraps the standard Go SQL connection.

## 12. Conversations

There are two conversation types:

- Private conversation: exactly two users.
- Group conversation: one or more users, with group name/photo.

### Starting a Private Conversation

Frontend:

- User searches another username.
- User clicks a result.
- Vue calls `POST /conversations` with:

```json
{ "userId": "target-user-id" }
```

Backend:

1. Checks target user is not the current user.
2. Checks target user exists.
3. Checks whether a private conversation already exists between the two users.
4. If yes, returns existing conversation.
5. If no, creates a new conversation and inserts both members.

Important SQL idea:

`GetPrivateConversation` joins `conversation_members` twice: once for user 1 and once for user 2. This finds a conversation that contains both users.

### Listing Conversations

Endpoint:

```text
GET /users/:userId/conversations
```

Security:

- The path `userId` must match the authenticated user ID.
- Otherwise the backend returns `403 Forbidden`.

Database:

- `GetUserConversations` gets all conversations for the user.
- It joins `conversations` with `conversation_members`.
- It uses a window function, `ROW_NUMBER() OVER (PARTITION BY conversation_id ORDER BY created_at DESC)`, to get the latest message for each conversation.

Professor may ask: "Why use a join?"

Answer:

> Conversations are spread across several tables. The join lets us combine membership, conversation metadata, and latest message information. Without joins, we would need many separate queries and the code would be slower and more complex.

### Opening a Conversation

Endpoint:

```text
GET /conversations/:conversationId
```

Backend:

1. Checks that the user is a member.
2. Loads conversation metadata.
3. Updates `last_read_at` for the current user.
4. Gets members.
5. Gets messages in reverse chronological order.
6. Adds sender username, checkmarks, comments, reply previews, and photo URLs to the response.

Important:

Opening a conversation marks it as read.

## 13. Messages

Messages are handled in `service/api/messages.go`.

### Sending Text

Endpoint:

```text
POST /conversations/:conversationId/messages
```

JSON body:

```json
{
  "type": "text",
  "content": "hello",
  "replyToId": "optional-message-id"
}
```

Backend:

1. Checks the user is a member of the conversation.
2. Creates a UUID for the message.
3. Reads JSON body.
4. Stores message in SQLite.
5. Returns a message response.

### Sending Photos

Photo messages use `multipart/form-data`.

Fields:

- `photo`: uploaded file.
- `type`: `"photo"`.
- `replyToId`: optional.

Photos are stored directly as BLOBs in SQLite.

The backend limits uploads to 5 MB.

Professor may ask: "Is storing images in the database ideal?"

Answer:

> It is simple for a course project because everything stays inside one SQLite file. In a production system, I would usually store images in object storage such as S3 and store only URLs or object keys in the database.

### Deleting Messages

Endpoint:

```text
DELETE /messages/:messageId
```

Only the sender can delete their own message. The backend:

1. Loads the message.
2. Checks `msg.SenderID == ctx.UserID`.
3. Deletes comments for that message.
4. Deletes the message.

### Forwarding Messages

Endpoint:

```text
POST /conversations/:conversationId/messages/forward
```

Body:

```json
{ "messageId": "original-message-id" }
```

Backend:

1. Checks user belongs to the target conversation.
2. Loads the original message.
3. Checks user belongs to the original message's conversation.
4. Creates a new message with the same content/photo.
5. Sets `Forwarded = true`.

Security detail:

If the user cannot access the original conversation, the backend returns "not found" behavior instead of exposing that the message exists.

## 14. Replies

Replies are stored using `reply_to_id` in the `messages` table.

When retrieving a conversation:

- If a message has `ReplyToID`, the backend loads the original message.
- It adds a preview object to the response.

This is a self-reference: a message row can point to another message row.

Professor may ask: "What is a self-referencing relationship?"

Answer:

> It is when a row in a table refers to another row in the same table. Here, `messages.reply_to_id` refers to `messages.id`.

## 15. Reactions and Comments

The code calls these comments, while the UI presents them like reactions.

Endpoint:

```text
PUT /messages/:messageId/comment
DELETE /messages/:messageId/comment
```

Rules:

- User must be a member of the message's conversation.
- Comment length must be 1 to 50 characters.
- One comment per user per message.
- In the UI, toggling reaction adds/removes a thumbs-up reaction.

Database:

```sql
PRIMARY KEY (message_id, user_id)
```

This prevents duplicate reactions from the same user on the same message.

## 16. Groups

Groups are just conversations with `type = 'group'`.

### Create Group

Endpoint:

```text
POST /groups
```

Body:

```json
{ "name": "Study group" }
```

The creator is automatically inserted into `conversation_members`.

### Add Member

Endpoint:

```text
POST /groups/:groupId/members
```

The requester must already be a group member. The target user must exist.

### Leave Group

Endpoint:

```text
DELETE /groups/:groupId/members/:userId
```

Important rule:

- Users can only remove themselves.
- There is no admin role in this implementation.

### Rename Group and Set Photo

Only group members can rename the group or upload a group photo.

Professor may ask: "Do groups have admins?"

Answer:

> No. The implementation uses simple membership only. Any group member can add members or rename the group. A production app would probably add roles like admin/owner.

## 17. Read Receipts and Checkmarks

Read receipt logic is in `GetMessageCheckmarks` in `service/database/messages.go`.

Return values:

- `0`: just sent. The comment mentions this state, but the code mostly returns 1 or 2.
- `1`: received.
- `2`: read by all other members.

How it works:

1. Get the message's conversation, sender, and creation time.
2. Count all conversation members except the sender.
3. Count how many of those members have `last_read_at >= message.created_at`.
4. If everyone has read, return 2.
5. Otherwise return 1.

When is `last_read_at` updated?

When a user opens a conversation through:

```text
GET /conversations/:conversationId
```

Good oral answer:

> The project stores a `last_read_at` timestamp per user per conversation. When a user opens a chat, that timestamp is updated. To know whether a message is read, the backend checks whether all other members have a `last_read_at` newer than the message creation time.

## 18. Frontend Architecture

The frontend starts in `webui/src/main.js`.

It:

1. Creates the Vue app.
2. Imports the root `App.vue`.
3. Installs Vue Router.
4. Registers Axios globally.
5. Mounts the app to the DOM.

### App.vue

`App.vue` only contains:

```html
<router-view />
```

This means the current route decides which page component is shown.

### Router

Defined in `webui/src/router/index.js`.

Routes:

```text
/                    LoginView
/conversations       ConversationsView
/conversations/:id   ChatView
/profile             ProfileView
```

The router has an auth guard:

- If a route requires auth and there is no token, redirect to `/`.
- If the user is already logged in and tries to open `/`, redirect to `/conversations`.

### Axios Service

Defined in `webui/src/services/axios.js`.

Important features:

- Base URL is `/api`.
- Timeout is 10 seconds.
- Request interceptor adds:

```text
Authorization: Bearer <token>
```

- Response interceptor handles `401` by clearing local storage and redirecting to login.

Good oral answer:

> Axios is centralized so components do not repeat authentication code. Every request automatically includes the Bearer token if it exists.

## 19. Main Frontend Screens

### LoginView

User enters username. The component calls:

```text
POST /session
```

On success, it stores token/user ID/username in local storage and navigates to `/conversations`.

### ConversationsView

Shows:

- Conversation list.
- Search users.
- Start private chat.
- Create group.
- Link to profile.

It calls:

```text
GET  /users/:userId/conversations
GET  /users?q=query
POST /conversations
POST /groups
```

It polls conversations every 5 seconds.

### ChatView

Shows:

- Messages.
- Text input.
- Photo upload.
- Reply.
- Reaction.
- Forward.
- Delete.
- Group settings.

It calls:

```text
GET    /conversations/:id
POST   /conversations/:id/messages
POST   /conversations/:id/messages/forward
DELETE /messages/:messageId
PUT    /messages/:messageId/comment
DELETE /messages/:messageId/comment
PUT    /groups/:id/name
POST   /groups/:id/members
DELETE /groups/:id/members/:userId
```

It polls the conversation every 3 seconds.

### ProfileView

Lets the user:

- Change username.
- Upload profile photo.
- Log out.

It calls:

```text
PUT /users/:userId/username
PUT /users/:userId/photo
```

## 20. Polling Instead of WebSockets

The project does not use WebSockets. It uses polling:

- Conversation list refreshes every 5 seconds.
- Open chat refreshes every 3 seconds.

Professor may ask: "Why not WebSockets?"

Answer:

> WebSockets are better for real-time messaging at scale, but they add complexity: persistent connections, connection lifecycle, reconnection logic, and server-side broadcasting. Polling is simpler and enough for this course project. The tradeoff is that polling is less efficient and messages are not truly instant.

## 21. Serving the Frontend from Go

The file `cmd/webapi/frontend.go` allows one Go server to serve both:

- API requests.
- Vue production files from `webui/dist`.

Logic:

1. If the path starts with `/api/`, strip `/api` and send it to the API router.
2. If the method is not GET, treat it as API.
3. If it is GET, check whether the requested file exists in `webui/dist`.
4. If the file exists, serve it.
5. If the request accepts HTML, serve `index.html`.
6. Otherwise pass to API router.

Why serve `index.html` for unknown frontend paths?

> Vue Router uses browser-side routing. If the user opens `/conversations/123` directly, there is no real file at that path. The server must send `index.html`, then Vue reads the URL and renders the correct page.

Security detail:

The code cleans the path and checks it does not escape `webui/dist`, to prevent directory traversal.

## 22. CORS

CORS is configured in `cmd/webapi/cors.go`.

It allows:

- All origins: `*`.
- Methods: GET, POST, OPTIONS, DELETE, PUT.
- Headers: Authorization, Content-Type.
- Max age: 1 second.

Why CORS is needed:

> During development, the frontend may run on Vite's port 5173 while the backend runs on 3000. Browsers block cross-origin requests unless the backend explicitly allows them.

## 23. Docker and Deployment

The main `Dockerfile` uses a multi-stage build.

### Stage 1: Frontend Builder

Uses `node:20-alpine`.

It:

1. Installs frontend dependencies.
2. Builds Vue.
3. Produces `webui/dist`.

### Stage 2: Backend Builder

Uses `golang:1.21`.

It:

1. Copies Go files and vendored dependencies.
2. Builds the Go binary with `go build -mod=vendor`.

### Stage 3: Runtime

Uses `debian:bookworm-slim`.

It:

1. Copies the backend binary.
2. Copies the built frontend files.
3. Creates `/app/data`.
4. Sets environment variables.
5. Runs `./webapi`.

Why multi-stage?

> The final image does not need Node, npm, Go compiler, or source build tools. Multi-stage Docker builds keep the runtime image smaller and cleaner.

Why Debian runtime?

> The SQLite Go driver uses CGO, which depends on C libraries. Debian is a safer base for CGO and SQLite than very minimal images.

### Docker Compose

`docker-compose.yml` runs one service:

```yaml
webapi:
  build: .
  ports:
    - "3000:3000"
  environment:
    - WASATEXT_DB_FILENAME=/app/data/wasatext.db
  volumes:
    - wasadata:/app/data
```

The volume keeps the SQLite database even if the container is recreated.

## 24. OpenAPI Contract

The API is documented in `doc/api.yaml`.

OpenAPI describes:

- Available endpoints.
- HTTP methods.
- Request bodies.
- Response codes.
- Schemas.
- Authentication.

Professor may ask: "Why is OpenAPI useful?"

Answer:

> OpenAPI acts as a contract between frontend and backend. It documents exactly what the frontend can call and what the backend returns. It also helps testing, documentation, and future code generation.

## 25. Important Request Lifecycles

### A. User Sends a Text Message

```text
User types message in ChatView
  -> Vue calls Axios POST /api/conversations/:id/messages
  -> Axios adds Authorization header
  -> FrontendHandler strips /api
  -> httprouter matches /conversations/:id/messages
  -> auth wrapper validates Bearer token
  -> sendMessage checks conversation membership
  -> sendMessage validates body and creates database.Message
  -> database.CreateMessage inserts into SQLite
  -> backend returns JSON message response
  -> Vue reloads conversation and updates UI
```

### B. User Opens a Conversation

```text
User clicks conversation
  -> Vue navigates to /conversations/:id
  -> ChatView calls GET /api/conversations/:id
  -> backend authenticates user
  -> backend checks membership
  -> backend marks conversation read
  -> backend loads members and messages
  -> backend calculates checkmarks and comments
  -> Vue displays messages
```

### C. User Creates a Group

```text
User enters group name
  -> Vue calls POST /api/groups
  -> backend authenticates user
  -> backend validates name length
  -> backend creates group conversation
  -> database transaction inserts conversation and creator membership
  -> backend returns group response
  -> Vue opens the group chat
```

### D. User Changes Username

```text
User enters new username in ProfileView
  -> Vue calls PUT /api/users/:userId/username
  -> backend checks path userId equals authenticated user
  -> backend validates regex
  -> backend checks uniqueness
  -> database updates users table
  -> frontend updates localStorage username
```

## 26. Validation and Error Handling

Examples of validation:

- Username must match `^[a-zA-Z0-9_]{3,16}$`.
- Group name must be 1 to 50 characters.
- Comment must be 1 to 50 characters.
- Photo body is limited to 5 MB.
- Users cannot start private conversations with themselves.
- Users can only update their own username/photo.
- Users can only delete their own messages.
- Users must be conversation members to read/send/comment/forward.

Common HTTP status codes:

- `201 Created`: login/session created, message created, group created.
- `204 No Content`: update/delete success with no response body.
- `400 Bad Request`: invalid body or invalid input.
- `401 Unauthorized`: missing or invalid Bearer token.
- `403 Forbidden`: authenticated but not allowed.
- `404 Not Found`: missing user/group/message/photo.
- `409 Conflict`: username already taken.
- `500 Internal Server Error`: unexpected database/server error.

## 27. Strengths of the Project

You can confidently mention:

- Clear separation between API and database layers.
- REST API documented by OpenAPI.
- Simple but complete messaging feature set.
- SQLite makes the app easy to run.
- Docker multi-stage build packages backend and frontend together.
- Auth wrapper prevents repeated auth code in each handler.
- Axios interceptor prevents repeated auth code in each frontend call.
- Database uses relationships and joins instead of storing everything in one table.
- Transactions are used when creating private/group conversations, so partial data is avoided.
- The frontend has route guards and centralized API handling.

## 28. Limitations and Honest Improvements

Professors often like when you can criticize your own project.

### Authentication is simplified

Current:

- Token is just user ID.
- No password.
- No expiration.

Production improvement:

- Password hashing.
- HTTPS.
- Signed JWT or secure HTTP-only cookies.
- Token expiration and refresh.

### Polling is simple but inefficient

Current:

- Chat polls every 3 seconds.
- Conversation list polls every 5 seconds.

Production improvement:

- WebSockets or Server-Sent Events for real-time updates.

### Images are stored in SQLite

Current:

- Photos are BLOBs in SQLite.

Production improvement:

- Store files in object storage or filesystem.
- Store only URLs/keys in database.

### No admin roles in groups

Current:

- Any group member can add members or rename the group.

Production improvement:

- Add owner/admin roles.
- Only admins can manage group settings.

### Content type handling could be more robust

Current:

- Text messages check if `Content-Type` equals exactly `application/json`.

Potential issue:

- Some clients send `application/json; charset=utf-8`.

Improvement:

- Parse content type using Go's `mime.ParseMediaType` or check prefix.

### Photo content type is simplified

Current:

- Photo endpoints always respond with `image/jpeg`.

Potential issue:

- Uploaded file might be PNG or another image type.

Improvement:

- Store MIME type alongside the image.

### No automated tests visible

Current:

- The repo does not show a dedicated test suite.

Improvement:

- Add unit tests for database methods.
- Add handler tests using `httptest`.
- Add frontend component or end-to-end tests.

### Healthcheck mismatch

Current:

- The API route `/liveness` returns `204 No Content`.
- The separate `cmd/healthcheck` utility checks for `200 OK`.

Improvement:

- Either change `/liveness` to return `200 OK`, or update the healthcheck utility to accept `204 No Content` as success.

### Frontend route nesting could be cleaner

Current:

- `ConversationsView` contains a `<router-view>`, but the router defines `/conversations` and `/conversations/:id` as sibling routes, not nested child routes.

Improvement:

- Either remove the unused nested `<router-view>` or define `/conversations/:id` as a child of `/conversations` if the intended design is a sidebar plus chat panel layout.

## 29. Likely Professor Questions and Strong Answers

### Q1. Explain the architecture of the project.

Answer:

> It is a full-stack layered monolith. The Vue frontend runs in the browser and communicates with the Go backend through REST APIs. The Go backend has an API layer for HTTP routing, authentication, validation, and JSON responses. It calls a database layer through an interface. The database layer stores data in SQLite. Docker builds the frontend and backend into one deployable container.

### Q2. What happens when a user logs in?

Answer:

> The frontend sends `POST /session` with a username. The backend validates the username format, checks if that username exists, and either returns the existing user ID or creates a new user with a UUID. The returned ID is stored in local storage and sent in future requests as a Bearer token.

### Q3. How does authentication work?

Answer:

> Most API routes are wrapped by an auth middleware. It reads the `Authorization` header, checks the `Bearer` format, treats the token as a user ID, and verifies that the user exists in the database. If valid, it passes a request context containing the user ID to the handler.

### Q4. Why is there a database interface?

Answer:

> The interface separates the API layer from the SQL implementation. API handlers call methods like `CreateMessage` or `GetConversationMessages` instead of writing SQL. This keeps responsibilities separate and makes the code easier to test or change later.

### Q5. How are private chats represented in the database?

Answer:

> A private chat is a row in `conversations` with type `private`. Its two users are connected through rows in `conversation_members`. The conversation itself does not store user IDs directly; membership is modeled through the junction table.

### Q6. How are groups represented?

Answer:

> Groups are also conversations, but with type `group` and a `group_name`. Members are stored in the same `conversation_members` table. This avoids duplicating message logic for private and group chats.

### Q7. Why use a junction table?

Answer:

> Because users and conversations have a many-to-many relationship. A user can be in many conversations, and a conversation can have many users. The junction table represents that relationship cleanly.

### Q8. How do read receipts work?

Answer:

> Each membership row has a `last_read_at` timestamp. When a user opens a conversation, the backend updates that timestamp. For each message, the backend checks if all other members have a `last_read_at` timestamp newer than the message creation time. If yes, the message is read by all.

### Q9. How do you prevent users from reading conversations they are not part of?

Answer:

> Before returning a conversation or allowing a message/comment, the backend calls `IsConversationMember`. If the authenticated user is not in the conversation, the backend returns `403 Forbidden`.

### Q10. Why use polling instead of WebSockets?

Answer:

> Polling is easier to implement and sufficient for a course project. WebSockets would provide better real-time behavior but require persistent connections, reconnection handling, and more server-side state.

### Q11. How does the frontend send authentication?

Answer:

> The Axios service has a request interceptor. Before every request, it reads `wasatext_token` from local storage and adds `Authorization: Bearer <token>`.

### Q12. What is the role of Vue Router?

Answer:

> Vue Router maps browser paths to Vue components. For example, `/conversations` renders the conversations list, and `/conversations/:id` renders the chat view. It also has a navigation guard that prevents unauthenticated users from opening protected pages.

### Q13. Why does the backend serve `index.html` for unknown GET paths?

Answer:

> Because the frontend is a single page application. Paths like `/conversations/123` are handled by Vue Router in the browser, not by real files on the server. The backend sends `index.html`, then Vue renders the correct page.

### Q14. How are photos handled?

Answer:

> Profile, group, and message photos are uploaded as binary data. The backend limits uploads to 5 MB and stores the bytes as BLOBs in SQLite. Photo retrieval endpoints return the stored bytes.

### Q15. Why use Docker?

Answer:

> Docker makes the project reproducible. The image contains the compiled Go backend, the built Vue frontend, and the runtime dependencies. Docker Compose maps port 3000 and creates a volume for the SQLite database.

### Q16. What is a multi-stage Docker build?

Answer:

> It uses separate build stages for frontend and backend, then copies only the final artifacts into a smaller runtime image. This avoids shipping Node, Go compiler, and source build tools in production.

### Q17. What is CORS and why do you need it?

Answer:

> CORS is a browser security mechanism that controls cross-origin HTTP requests. During development, the frontend may run on a different port than the backend, so the backend must allow those requests and headers.

### Q18. How is a forwarded message implemented?

Answer:

> The backend loads the original message, checks that the user can access it, then creates a new message in the target conversation with the same content or photo and sets the forwarded flag.

### Q19. How is message deletion controlled?

Answer:

> The backend loads the message and checks whether the authenticated user is the sender. Only the sender can delete it.

### Q20. What would you improve if you had more time?

Answer:

> I would improve authentication with passwords and signed tokens, replace polling with WebSockets, store images outside the database, add group admin roles, store MIME types for images, and add automated tests for handlers and database methods.

## 30. Code Files You Should Know by Heart

### `cmd/webapi/main.go`

Know this as the startup file.

Say:

> This file wires everything together: configuration, database connection, API router, frontend serving, CORS, HTTP server, and graceful shutdown.

### `service/api/api-handler.go`

Know this as the route map.

Say:

> This file connects URLs and HTTP methods to handler functions. It also shows which endpoints require authentication because most routes use `rt.wrap`.

### `service/api/auth.go`

Know this as authentication middleware.

Say:

> Instead of repeating token validation inside every handler, the wrapper handles it once and passes the authenticated user ID through `RequestContext`.

### `service/database/database.go`

Know this as schema and database interface.

Say:

> This file defines the database contract, data models, and table creation. It is the boundary between business logic and SQL.

### `webui/src/services/axios.js`

Know this as the frontend API client.

Say:

> It centralizes API calls, sets the `/api` base URL, adds the Bearer token automatically, and handles 401 responses.

### `webui/src/router/index.js`

Know this as frontend navigation and route protection.

Say:

> It maps URLs to components and prevents access to protected pages when no token exists.

## 31. Small Details That Make You Sound Prepared

- The backend uses `httprouter`, not Gorilla Mux.
- Gorilla is used only through `gorilla/handlers` for CORS.
- User IDs and message IDs are UUID strings.
- SQLite tables are created automatically with `CREATE TABLE IF NOT EXISTS`.
- Private and group chats share the same `conversations` and `messages` tables.
- `conversation_members` is the key table for permissions and read receipts.
- Photos are stored as BLOBs.
- The frontend sends API requests to `/api`, but `FrontendHandler` strips `/api` before routing to the backend.
- The frontend has two polling intervals: 5 seconds for conversation list and 3 seconds for open chat.
- The app uses local storage, which is easy but less secure than HTTP-only cookies.
- Docker Compose uses a named volume called `wasadata` to persist SQLite data.

## 32. Suggested Oral Presentation Order

Use this order if the professor says, "Explain the project."

1. Start with the goal: WhatsApp-like messaging app.
2. Explain the three main parts: Vue, Go, SQLite.
3. Explain layered monolith architecture.
4. Walk through login and authentication.
5. Explain database schema.
6. Explain conversations and messages.
7. Explain groups, replies, forwarding, comments.
8. Explain frontend views and Axios.
9. Explain Docker deployment.
10. End with limitations and improvements.

Short version:

> WASAText is a Vue + Go + SQLite messaging application. The frontend is a single page app. It calls a Go REST API using Axios. The backend authenticates users with a simplified Bearer token, validates permissions, and calls a database layer through an interface. SQLite stores users, conversations, members, messages, and comments. Docker builds the frontend and backend into one container and persists the SQLite file using a volume.

## 33. Final Cheat Sheet

```text
Frontend: Vue 3, Vue Router, Axios, Vite
Backend: Go, net/http, httprouter, logrus
Database: SQLite through database/sql and go-sqlite3
Auth: Bearer token = user UUID
Storage: users, conversations, conversation_members, messages, message_comments
Real-time strategy: polling, not WebSockets
Deployment: Docker multi-stage build + docker-compose volume
API docs: doc/api.yaml OpenAPI
Main backend entry: cmd/webapi/main.go
Route map: service/api/api-handler.go
Schema: service/database/database.go
Frontend API client: webui/src/services/axios.js
```

Most important sentence:

> The central design idea is separation of concerns: Vue handles UI, API handlers handle HTTP and permissions, the database layer handles SQL, and Docker handles packaging.
