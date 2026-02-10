# Module 7: DevOps & Deployment (Docker)

This final module covers how we package and run our application anywhere.

## 7.1 Docker Philosophy
"It works on my machine" is not acceptable.
Docker solves this by packaging the code AND its dependencies (OS libraries, runtime, database drivers) into a single "Image".

## 7.2 The `Dockerfile` Anatomy
We use a **Multi-Stage Build**. This is an advanced technique.

### Stage 1: `frontend-builder`
- Base: `node:20-alpine` (Small Node.js image).
- Goal: Build Vue.js.
- Commands: `npm install`, `npm run build`.
- Output: A `dist/` folder with compiled files.

### Stage 2: `backend-builder`
- Base: `golang:1.21`.
- Why Debian?: Alpine Linux uses `musl` libc. Standard Linux uses `glibc`. Go's `sqlite3` driver requires CGO (C bindings). Compiling CGO on Alpine can be tricky (often fails with `pread64` errors). Debian is more stable for this.
- Commands: `go build`.
- Output: A single binary `webapi`.

### Stage 3: The Runtime Image
- Base: `debian:bookworm-slim`.
- Why Slim?: We don't need Go compiler or Node in production.
- Action:
    1.  Copy binary from Stage 2.
    2.  Copy frontend files from Stage 1 into `webui/dist`.
- Result: A tiny image tailored for running the app.

## 7.3 Running the Container
```bash
docker run -p 3000:3000 wasatext
```
- **-p 3000:3000**: Maps port 3000 on your laptop (Host) to port 3000 inside the container.
- **Data Persistence**: By default, SQLite data is lost when container restarts.
    - **Fix**: Use Volume Mounting. `docker run -v $(pwd)/data:/app/data ...`
    - This maps a folder on your laptop to `/app/data` inside container, keeping `wasatext.db` safe.

## 7.4 Summary & Oral Exam Tips
- **Q**: "Why multi-stage?" -> **A**: To separate build tools (heavy) from runtime (light).
- **Q**: "Why SQLite?" -> **A**: Embedded, zero-config, portable.
- **Q**: "Why Debian?" -> **A**: Reliable CGO support for SQLite compared to Alpine.
