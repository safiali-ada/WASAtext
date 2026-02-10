package database

import (
	"database/sql"
	"errors"
)

// CreateMessage creates a new message
func (db *appdbimpl) CreateMessage(msg *Message) error {
	forwarded := 0
	if msg.Forwarded {
		forwarded = 1
	}

	var replyToID interface{}
	if msg.ReplyToID != "" {
		replyToID = msg.ReplyToID
	}

	_, err := db.c.Exec(`
		INSERT INTO messages (id, conversation_id, sender_id, content, photo, type, reply_to_id, forwarded, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, msg.ID, msg.ConversationID, msg.SenderID, msg.Content, msg.Photo, msg.Type, replyToID, forwarded)
	return err
}

// GetMessage retrieves a message by ID
func (db *appdbimpl) GetMessage(id string) (*Message, error) {
	var msg Message
	var replyToID sql.NullString
	var forwarded int

	err := db.c.QueryRow(`
		SELECT id, conversation_id, sender_id, content, photo, type, reply_to_id, forwarded, created_at
		FROM messages WHERE id = ?
	`, id).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Photo, &msg.Type, &replyToID, &forwarded, &msg.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if replyToID.Valid {
		msg.ReplyToID = replyToID.String
	}
	msg.Forwarded = forwarded == 1

	return &msg, nil
}

// DeleteMessage deletes a message
func (db *appdbimpl) DeleteMessage(id string) error {
	// First delete comments on the message
	_, err := db.c.Exec("DELETE FROM message_comments WHERE message_id = ?", id)
	if err != nil {
		return err
	}

	result, err := db.c.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("message not found")
	}
	return nil
}

// GetConversationMessages retrieves all messages in a conversation (reverse chronological)
func (db *appdbimpl) GetConversationMessages(conversationID string) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT id, conversation_id, sender_id, content, photo, type, reply_to_id, forwarded, created_at
		FROM messages 
		WHERE conversation_id = ?
		ORDER BY created_at DESC
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var replyToID sql.NullString
		var forwarded int

		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Photo, &msg.Type, &replyToID, &forwarded, &msg.CreatedAt); err != nil {
			return nil, err
		}

		if replyToID.Valid {
			msg.ReplyToID = replyToID.String
		}
		msg.Forwarded = forwarded == 1

		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// GetMessageCheckmarks calculates checkmarks for a message
// 0 = just sent, 1 = received by all, 2 = read by all
func (db *appdbimpl) GetMessageCheckmarks(messageID string) (int, error) {
	// Get the message info
	var conversationID, senderID, createdAt string
	err := db.c.QueryRow("SELECT conversation_id, sender_id, created_at FROM messages WHERE id = ?", messageID).
		Scan(&conversationID, &senderID, &createdAt)
	if err != nil {
		return 0, err
	}

	// Count total members (excluding sender)
	var totalMembers int
	err = db.c.QueryRow("SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id != ?",
		conversationID, senderID).Scan(&totalMembers)
	if err != nil {
		return 0, err
	}

	if totalMembers == 0 {
		return 2, nil // No other members, consider it read
	}

	// Count members who have read (last_read_at >= message created_at)
	var readCount int
	err = db.c.QueryRow(`
		SELECT COUNT(*) FROM conversation_members 
		WHERE conversation_id = ? AND user_id != ? AND last_read_at >= ?
	`, conversationID, senderID, createdAt).Scan(&readCount)
	if err != nil {
		return 0, err
	}

	if readCount == totalMembers {
		return 2, nil // All have read
	}
	return 1, nil // Message exists = received by all (simplification)
}

// AddComment adds a comment/reaction to a message
func (db *appdbimpl) AddComment(messageID, userID, comment string) error {
	_, err := db.c.Exec(`
		INSERT OR REPLACE INTO message_comments (message_id, user_id, comment)
		VALUES (?, ?, ?)
	`, messageID, userID, comment)
	return err
}

// RemoveComment removes a comment from a message
func (db *appdbimpl) RemoveComment(messageID, userID string) error {
	result, err := db.c.Exec("DELETE FROM message_comments WHERE message_id = ? AND user_id = ?", messageID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("comment not found")
	}
	return nil
}

// GetMessageComments gets all comments on a message
func (db *appdbimpl) GetMessageComments(messageID string) ([]Comment, error) {
	rows, err := db.c.Query(`
		SELECT mc.message_id, mc.user_id, u.username, mc.comment
		FROM message_comments mc
		INNER JOIN users u ON mc.user_id = u.id
		WHERE mc.message_id = ?
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.MessageID, &c.UserID, &c.Username, &c.Comment); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
