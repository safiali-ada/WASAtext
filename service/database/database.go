/*
Package database is the middleware between the app database and the code.
All data (de)serialization (save/load) from a persistent database are handled here.
Database specific logic should never escape this package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	// User operations
	CreateUser(id, username string) error
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUsername(userID, newUsername string) error
	UpdateUserPhoto(userID string, photo []byte) error
	SearchUsers(query string) ([]User, error)

	// Conversation operations
	CreatePrivateConversation(id, user1ID, user2ID string) error
	CreateGroupConversation(id, name, creatorID string) error
	GetConversation(id string) (*Conversation, error)
	GetUserConversations(userID string) ([]ConversationPreview, error)
	GetPrivateConversation(user1ID, user2ID string) (*Conversation, error)
	IsConversationMember(conversationID, userID string) (bool, error)
	MarkConversationRead(conversationID, userID string) error

	// Group operations
	AddGroupMember(groupID, userID string) error
	RemoveGroupMember(groupID, userID string) error
	UpdateGroupName(groupID, name string) error
	UpdateGroupPhoto(groupID string, photo []byte) error
	GetGroupMembers(groupID string) ([]User, error)

	// Message operations
	CreateMessage(msg *Message) error
	GetMessage(id string) (*Message, error)
	DeleteMessage(id string) error
	GetConversationMessages(conversationID string) ([]Message, error)
	GetMessageCheckmarks(messageID string) (int, error)

	// Comment operations
	AddComment(messageID, userID, comment string) error
	RemoveComment(messageID, userID string) error
	GetMessageComments(messageID string) ([]Comment, error)

	Ping() error
}

// User represents a user in the database
type User struct {
	ID       string
	Username string
	Photo    []byte
}

// Conversation represents a conversation
type Conversation struct {
	ID        string
	Type      string // "private" or "group"
	GroupName string
	Photo     []byte
}

// ConversationPreview represents a conversation in the list
type ConversationPreview struct {
	ID            string
	Type          string
	Name          string
	Photo         []byte
	LatestMessage *MessagePreview
}

// MessagePreview represents a message preview
type MessagePreview struct {
	Content   string
	Timestamp string
	SenderID  string
}

// Message represents a message
type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	Content        string
	Photo          []byte
	Type           string // "text" or "photo"
	ReplyToID      string
	Forwarded      bool
	CreatedAt      string
}

// Comment represents a reaction/comment on a message
type Comment struct {
	MessageID string
	UserID    string
	Username  string
	Comment   string
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Create tables if they don't exist
	err := createTables(db)
	if err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func createTables(db *sql.DB) error {
	sqlStmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			photo BLOB
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			group_name TEXT,
			photo BLOB
		)`,
		`CREATE TABLE IF NOT EXISTS conversation_members (
			conversation_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			last_read_at DATETIME,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
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
		)`,
		`CREATE TABLE IF NOT EXISTS message_comments (
			message_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			comment TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY (message_id) REFERENCES messages(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, stmt := range sqlStmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
