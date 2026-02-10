package database

import (
	"database/sql"
	"errors"
)

// CreatePrivateConversation creates a private conversation between two users
func (db *appdbimpl) CreatePrivateConversation(id, user1ID, user2ID string) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO conversations (id, type) VALUES (?, 'private')", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?), (?, ?)",
		id, user1ID, id, user2ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreateGroupConversation creates a group conversation
func (db *appdbimpl) CreateGroupConversation(id, name, creatorID string) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO conversations (id, type, group_name) VALUES (?, 'group', ?)", id, name)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)", id, creatorID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetConversation retrieves a conversation by ID
func (db *appdbimpl) GetConversation(id string) (*Conversation, error) {
	var conv Conversation
	err := db.c.QueryRow("SELECT id, type, group_name, photo FROM conversations WHERE id = ?", id).
		Scan(&conv.ID, &conv.Type, &conv.GroupName, &conv.Photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetPrivateConversation finds an existing private conversation between two users
func (db *appdbimpl) GetPrivateConversation(user1ID, user2ID string) (*Conversation, error) {
	query := `
		SELECT c.id, c.type, c.group_name, c.photo 
		FROM conversations c
		INNER JOIN conversation_members cm1 ON c.id = cm1.conversation_id AND cm1.user_id = ?
		INNER JOIN conversation_members cm2 ON c.id = cm2.conversation_id AND cm2.user_id = ?
		WHERE c.type = 'private'
	`
	var conv Conversation
	err := db.c.QueryRow(query, user1ID, user2ID).Scan(&conv.ID, &conv.Type, &conv.GroupName, &conv.Photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetUserConversations retrieves all conversations for a user
func (db *appdbimpl) GetUserConversations(userID string) ([]ConversationPreview, error) {
	query := `
		SELECT c.id, c.type, c.group_name, c.photo,
			COALESCE(m.content, '') as latest_content,
			COALESCE(m.created_at, '') as latest_timestamp,
			COALESCE(m.sender_id, '') as latest_sender
		FROM conversations c
		INNER JOIN conversation_members cm ON c.id = cm.conversation_id AND cm.user_id = ?
		LEFT JOIN (
			SELECT conversation_id, content, created_at, sender_id,
				ROW_NUMBER() OVER (PARTITION BY conversation_id ORDER BY created_at DESC) as rn
			FROM messages
		) m ON c.id = m.conversation_id AND m.rn = 1
		ORDER BY m.created_at DESC NULLS LAST
	`
	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var previews []ConversationPreview
	for rows.Next() {
		var p ConversationPreview
		var groupName sql.NullString
		var latestContent, latestTimestamp, latestSender string

		if err := rows.Scan(&p.ID, &p.Type, &groupName, &p.Photo,
			&latestContent, &latestTimestamp, &latestSender); err != nil {
			return nil, err
		}

		if p.Type == "group" && groupName.Valid {
			p.Name = groupName.String
		}

		if latestContent != "" || latestTimestamp != "" {
			p.LatestMessage = &MessagePreview{
				Content:   latestContent,
				Timestamp: latestTimestamp,
				SenderID:  latestSender,
			}
		}

		previews = append(previews, p)
	}

	// For private conversations, we need to get the other user's name
	for i := range previews {
		if previews[i].Type == "private" {
			otherUser, err := db.getOtherUserInConversation(previews[i].ID, userID)
			if err == nil && otherUser != nil {
				previews[i].Name = otherUser.Username
			}
		}
	}

	return previews, rows.Err()
}

func (db *appdbimpl) getOtherUserInConversation(conversationID, userID string) (*User, error) {
	var user User
	err := db.c.QueryRow(`
		SELECT u.id, u.username, u.photo 
		FROM users u
		INNER JOIN conversation_members cm ON u.id = cm.user_id
		WHERE cm.conversation_id = ? AND u.id != ?
	`, conversationID, userID).Scan(&user.ID, &user.Username, &user.Photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IsConversationMember checks if a user is a member of a conversation
func (db *appdbimpl) IsConversationMember(conversationID, userID string) (bool, error) {
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id = ?",
		conversationID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkConversationRead marks all messages in a conversation as read for a user
func (db *appdbimpl) MarkConversationRead(conversationID, userID string) error {
	_, err := db.c.Exec("UPDATE conversation_members SET last_read_at = CURRENT_TIMESTAMP WHERE conversation_id = ? AND user_id = ?",
		conversationID, userID)
	return err
}

// AddGroupMember adds a user to a group
func (db *appdbimpl) AddGroupMember(groupID, userID string) error {
	_, err := db.c.Exec("INSERT OR IGNORE INTO conversation_members (conversation_id, user_id) VALUES (?, ?)", groupID, userID)
	return err
}

// RemoveGroupMember removes a user from a group
func (db *appdbimpl) RemoveGroupMember(groupID, userID string) error {
	result, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", groupID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not a member of group")
	}
	return nil
}

// UpdateGroupName updates the name of a group
func (db *appdbimpl) UpdateGroupName(groupID, name string) error {
	_, err := db.c.Exec("UPDATE conversations SET group_name = ? WHERE id = ? AND type = 'group'", name, groupID)
	return err
}

// UpdateGroupPhoto updates the photo of a group
func (db *appdbimpl) UpdateGroupPhoto(groupID string, photo []byte) error {
	_, err := db.c.Exec("UPDATE conversations SET photo = ? WHERE id = ? AND type = 'group'", photo, groupID)
	return err
}

// GetGroupMembers gets all members of a group
func (db *appdbimpl) GetGroupMembers(groupID string) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.photo 
		FROM users u
		INNER JOIN conversation_members cm ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Photo); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
