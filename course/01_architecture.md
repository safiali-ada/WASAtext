# Module 1: Architecture & Project Philosophy

Welcome to the **WASAText Implementation Course**. This module breaks down the high-level decisions that shape the entire project. Before looking at a single line of code, you must understand *why* the project is structured this way.

## 1.1 The Problem Statement
We need to build a "WhatsApp-like" messaging application with:
- **Real-time capabilities** (sending/receiving messages instantly).
- **User Identity** (simplified login, profiles).
- **Group Chats** (many-to-many communication).
- **Media Support** (sending photos).
- **Portability** (must run anywhere via Docker).

## 1.2 The Architecture: Monolith vs. Microservices
For a project of this scale (academic/startup MVP), a **Monolithic Architecture** is the correct choice.
- **Microservices**: Would require separate services for Auth, Messaging, User Profile, etc., all communicating via gRPC/HTTP. This adds huge complexity (service discovery, distributed tracing) with little benefit for a single team.
- **Monolith**: A single executable contains all logic.
    - **Pros**: Easy to deploy (one binary), easy to debug (one process), zero network latency between internal components.
    - **Our Twist**: We use a **Layered Monolith**. Inside the single binary, code is strictly separated into `api` (HTTP handling) and `database` (Data handling). This keeps the code clean and allows for future splitting if needed.

## 1.3 The Tech Stack Selection
Why did we choose these specific technologies?

### **Backend: Go (Golang)**
- **Concurrency**: Go's *goroutines* are lightweight threads managed by the Go runtime, not the OS. A typical server can handle 10,000+ concurrent connections (users) with minimal RAM.
- **Standard Library**: Go's `net/http` is production-grade. We don't need heavy frameworks (like Spring Boot or Django) just to handle HTTP requests.
- **Deployment**: Go compiles to a **single static binary**. No interpreter (Python), no JVM (Java), no runtime dependencies. This makes Docker images tiny (tens of MBs, not hundreds).

### **Frontend: Vue.js (SPA)**
- **Single Page Application (SPA)**: The user loads `index.html` *once*. After that, JavaScript intercepts all clicks and fetches JSON data from the backend. The page never reloads. This provides a "native app" feel.
- **Reactivity System**: Vue automatically updates the HTML when JavaScript variables change. If `messages.push(newMsg)` happens, the chat list updates instantly. No manual DOM manipulation (`document.getElementById`) is needed.

### **Database: SQLite**
- **Embedded**: SQLite is a C library that reads/writes directly to a disk file (`wasatext.db`). It is *not* a separate process listening on a port (like MySQL:3306).
- **Zero Configuration**: No user management, no permissions, no network config. Just a file.
- **ACID Compliant**: Despite its simplicity, it guarantees data integrity.
- **Why?**: For a portable assignment, requiring a separate PostgreSQL container adds complexity (docker-compose). SQLite keeps the architecture "self-contained".

## 1.4 The Project Structure (Standard Go Layout)
We follow the industry-standard Go project layout:

```text
/
├── cmd/
│   └── webapi/       # Main applications.
│       └── main.go   # The Entry Point. Kicks off the server.
├── service/
│   ├── api/          # The API Interface Layer (Controllers).
│   └── database/     # The Database Layer (Repositories).
├── webui/            # The Frontend source code (Node.js/Vue).
├── doc/              # Documentation (OpenAPI spec).
└── Dockerfile        # Blueprint for our deployment container.
```

### Key Takeaway
We strictly separate **Interface** (API) from **Implementation** (Database). The API handlers don't know *how* `GetUser` works (SQL? Mongo? File?), they just call an interface method. This is **Dependency Inversion principle**.
