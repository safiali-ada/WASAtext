# Module 2: The API Contract (OpenAPI)

In this module, we explore the **Contract-First** design philosophy. Before writing any Go code, we defined *exactly* how the API behaves. This is crucial for separating Backend and Frontend development.

## 2.1 What is OpenAPI?
OpenAPI (formerly Swagger) is a standard for describing REST APIs. It uses a YAML file to define:
- **Paths**: URLs available (e.g., `/session`).
- **Methods**: HTTP Verbs (GET, POST, PUT, DELETE).
- **Parameters**: Inputs (Query strings, Path variables).
- **Responses**: What the server sends back (200 OK, 400 Bad Request).
- **Schemas**: The shape of the data (JSON fields).

## 2.2 Analyzing `doc/api.yaml`
Our file `doc/api.yaml` is the source of truth.

### The Security Scheme
We use **Bearer Authentication**.
```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
```
- **Meaning**: Clients must send an HTTP Header: `Authorization: Bearer <token>`.
- **Implementation**: The backend validates this token. If missing/invalid, it returns `401 Unauthorized`.

### Key Endpoints Breakdown

#### 1. Login (`/session`)
- **Method**: `POST`
- **operationId**: `doLogin`
- **Goal**: Create a session.
- **Input**: `{ "name": "alice" }`
- **Output**: `{ "identifier": "uuid-..." }`
- **Why POST?**: We are *creating* a resource (a session). Even though it's simplified (no password), semantically it's a creation event.

#### 2. Send Message (`/conversations/{conversationId}/messages`)
- **Method**: `POST`
- **operationId**: `sendMessage`
- **Content-Type**: defined as `multipart/form-data` OR `application/json`.
    - *Why?* To support both text (JSON) and images (Multipart) in a single endpoint design.
- **Path Parameter**: `{conversationId}`. This tells the server *where* to post the message.

#### 3. Stream Messages (`/conversations/{conversationId}`)
- **Method**: `GET`
- **operationId**: `getConversation`
- **Why not WebSockets?**: For this project scale, **Short Polling** (client requests every 5s) is simpler and sufficient. The OpenAPI spec doesn't mandate the transport mechanism (Socket vs HTTP), but the design implies request-response.

## 2.3 Operation IDs
The `operationId` field (e.g., `doLogin`) is critical.
- It links the Abstract Spec to the Concrete Code.
- In `api-handler.go`, we map these IDs to Go functions:
  ```go
  // "doLogin" in YAML -> rt.doLogin function in Go
  rt.router.POST("/session", rt.doLogin)
  ```

## 2.4 Data Models (Schemas)
Defining schemas ensures both Backend and Frontend agree on data types.
- **User**: `id` (string), `username` (string), `photo` (string/binary).
- **Conversation**: `id`, `type` (group/private), `last_message`.
- **Message**: `id`, `sender_id`, `content`, `created_at`.

If the Backend sends `user_name` but the Spec says `username`, it's a bug. The Spec rules supreme.
