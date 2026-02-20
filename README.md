# WASAText 💬

A modern messaging platform built for the Web and Software Architecture course. Think WhatsApp meets simplicity.

## What's This?

WASAText is a full-stack messaging app where you can:
- Chat with friends in private conversations
- Create group chats
- Send photos and messages
- React to messages with comments
- Forward messages between chats

Built with Go (backend), Vue.js (frontend), and SQLite (database).

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (if running without Docker)
- Node 20+ (if developing the frontend)

### Running with Docker (Easiest!)

```bash
# Build and start everything
docker-compose up --build

# The app will be available at http://localhost:3000
```

That's it! Docker handles the frontend build, backend compilation, and database setup.

### Running Locally (For Development)

**Backend:**
```bash
go run ./cmd/webapi
```

**Frontend:**
```bash
cd webui
npm install  # or yarn install
npm run dev  # or yarn dev
```

## Project Structure

```
├── cmd/
│   ├── webapi/       # Main API server
│   └── healthcheck/  # Health check utility
├── service/
│   ├── api/          # API handlers
│   ├── database/     # Database logic
│   └── globaltime/   # Time utilities (for testing)
├── webui/            # Vue.js frontend
├── doc/              # API documentation (OpenAPI spec)
├── demo/             # Example config files
└── vendor/           # Go dependencies (vendored)
```

## API Documentation

Check out `doc/api.yaml` for the full OpenAPI specification. You can view it in tools like Swagger Editor.

Key endpoints:
- `POST /session` - Login/create user
- `GET /users/{userId}/conversations` - Get your chats
- `POST /conversations/{conversationId}/messages` - Send a message
- `POST /groups` - Create a group chat

## Development Tips

### Frontend Development
Use the included Node container for a clean environment:
```bash
./open-node.sh
```

### Go Vendoring
This project uses Go vendoring. After adding dependencies:
```bash
go mod tidy
go mod vendor
git add vendor/
```

### Database
SQLite database is stored in `/app/data/wasatext.db` (in Docker) or `./wasatext.db` (locally).

## What's Under the Hood?

- **Backend**: Go with Gorilla Mux for routing
- **Frontend**: Vue.js 3 with Vue Router
- **Database**: SQLite3 (simple and effective)
- **Containerization**: Docker with multi-stage builds

## Need Help?

Check the `/doc` folder for detailed API specs or dive into the code - it's well-commented!

---

Built with ☕ for the Web and Software Architecture course.
