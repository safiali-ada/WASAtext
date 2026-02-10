# Module 3: Backend Core Implementation

This module covers the "spinal cord" of our application: `cmd/webapi/main.go`. This is where the application bootstraps itself.

## 3.1 The Entry Point (`main.go`)
In Go, `func main()` is the first function executed.

### Step 1: Loading Configuration
Before anything else, we read environment variables.
- Function: `loadConfiguration()` (in `load-configuration.go`).
- **Why Env Vars?**: Keeping config separate from code allows us to deploy the same binary to different environments (local, dev, prod) just by changing `WASATEXT_DB_FILENAME`.
- **Implementation**: We use `os.Getenv` directly. Simple, dependency-free.

### Step 2: Database Connection
We connect to SQLite.
```go
db, err := sql.Open("sqlite3", cfg.DB.Filename)
```
- This opens the file (or creates it).
- We then wrap this raw connection in our custom `database` package (`database.New(db)`).
- **Migration Logic**: Inside `database.New`, we run `CREATE TABLE IF NOT EXISTS`. This ensures the schema is always up-to-date.

### Step 3: Router Initialization
We create the API router.
```go
apirouter, err := api.New(api.Config{...})
router := apirouter.Handler()
```
- `api.New` injects dependencies (DB connection, Logger) into the router. This is **Dependency Injection**. It makes testing easier (we could inject a mock DB).

### Step 4: The Unified Serving Trick (`frontend.go`)
This is the most critical part for our Docker deployment.
We wrap the API router with `FrontendHandler(router)`.

**What does `FrontendHandler` do?**
1.  **Intercepts Requests**: Every HTTP request hits this handler first.
2.  **API Check**: If the path starts with `/api/` or is a known API method (POST/PUT/DELETE), it passes the request to the `apiHandler` (our Go router).
3.  **Static File Check**: If not API, it checks if a file exists in `webui/dist` (the built Frontend). e.g., `/assets/index.js`. If yes, it serves the file.
4.  **SPA Routing**: If it's a browser navigation (e.g., `/conversations/123`), the file doesn't exist on disk. But we serve `index.html` anyway.
    - **Why?**: Because Vue Router (in the browser) sees the URL and knows what to render. The backend just needs to deliver the HTML shell.

### Step 5: Start Server
Finally, `http.ListenAndServe(cfg.Web.APIHost, finalHandler)` starts the server loop. It blocks forever, waiting for incoming TCP connections on port 3000.

## 3.2 CORS (Cross-Origin Resource Sharing)
Implemented in `cors.go`.
- **Problem**: Browsers block cross-origin requests (e.g., from `localhost:5173` to `localhost:3000`) for security.
- **Solution**: We add headers:
  - `Access-Control-Allow-Origin: *` (Allow anyone)
  - `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
  - `Access-Control-Max-Age: 1` (As per project requirements)
- **Why Max-Age 1?**: It forces the browser to re-check permissions frequently (almost every request), which is inefficient but required by the spec.

## 3.3 Folder Structure Deep Dive
- `cmd/webapi/`: Application entry point.
    - `main.go`: Bootstrapper.
    - `frontend.go`: Static file server logic.
    - `cors.go`: CORS middleware.
    - `load-configuration.go`: Config loader.
