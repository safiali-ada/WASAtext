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

# The frontend will be available at http://localhost:8081
# The backend API will be available at localhost:3000
```

That's it! Docker Compose builds `Dockerfile.frontend` and `Dockerfile.backend`. The browser connects directly to the backend on port 3000, matching the course's WASA Fountains example.

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

### Running for the homework evaluation

**Backend** (Go 1.25.1):
```bash
export GOTOOLCHAIN=go1.25.1
go mod tidy
go mod vendor
go run ./cmd/webapi
```

**Frontend** (inside the Node container):
```bash
./open-node.sh
yarn install
yarn run vite build --mode production
yarn run vite preview --host 0.0.0.0 --port 4173 --strictPort
```

Open `http://localhost:4173`, or `http://<your-LAN-IP>:4173` from another device on the same network.
The backend URL is the `__API_URL__` constant in `webui/vite.config.js` (default `http://localhost:3000`).
For the multi-device test you can replace `localhost` with your LAN IP as the checklist describes;
if you leave it unchanged, a page opened through the LAN IP automatically uses that IP for the backend.

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
SQLite uses `/tmp/wasatext.db` in the standalone backend container and `./wasatext.db` locally.

## What's Under the Hood?

- **Backend**: Go with `httprouter` for routing
- **Frontend**: Vue.js 3 with Vue Router
- **Database**: SQLite3 (simple and effective)
- **Containerization**: Docker with multi-stage builds

## Need Help?

Check the `/doc` folder for detailed API specs or dive into the code - it's well-commented!

---

Built with ☕ for the Web and Software Architecture course.
