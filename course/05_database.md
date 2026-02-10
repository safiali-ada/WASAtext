# Module 5: Database Design & Implementation

In this module, we explore the data layer. We use **SQLite**, an embedded SQL database.

## 5.1 Schema Design (`database.go`)
We define our tables using `CREATE TABLE IF NOT EXISTS` statements.

### Users Table
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    photo BLOB
);
```
- **ID**: UUID string.
- **Username**: Must be unique.
- **Photo**: Stored as BLOB (Binary Large Object). In a real-world app, we'd store the image in S3 and save the URL here, but for this project, storing directly in DB is simpler.

### Conversations Table
```sql
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    group_name TEXT,
    photo BLOB
);
```
- **Type**: 'private' or 'group'.
- **Group Name**: Only relevant if type='group'.

### Junction Table: `conversation_members`
This implements the **Many-to-Many Relationship**.
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
- **Composite Primary Key**: Ensures a user cannot join the same group twice.
- **last_read_at**: Crucial for "Blue Ticks" (Read Receipts). When Alice views the chat, we update this timestamp.

### Messages Table
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
- **reply_to_id**: Self-Referencing Foreign Key. Points to another message `id`.
- **forwarded**: Boolean flag (0/1 in SQLite).

## 5.2 Key SQL Operations

### `GetUserConversations`
This query joins `conversation_members` (to find chats user is in) with `conversations` (to get details) and `messages` (to get snippet of last message).
- **SQL JOIN**: Essential for efficiency. Instead of N+1 queries (looping), we get all data in 1 query.

### `GetMessageCheckmarks` (Logic)
How do we know if a message is read?
- **Status 0 (Sent)**: Default.
- **Status 1 (Received)**: Usually handled by server ACK (always true here).
- **Status 2 (Read)**:
    - We check `last_read_at` for ALL members in the group (excluding sender).
    - If `everyone's last_read_at > message.created_at`, then everyone has seen it.
    - Status = 2.
    - Else Status = 1.

## 5.3 Transactions
We use single operations here mostly, but for complex actions (like creating a group + adding members), we could use transactions (`BEGIN`, `COMMIT`). SQLite handles single statements atomically by default.
