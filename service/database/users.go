package database

import (
	"database/sql"
	"errors"
)

// CreateUser creates a new user
func (db *appdbimpl) CreateUser(id, username string) error {
	_, err := db.c.Exec("INSERT INTO users (id, username) VALUES (?, ?)", id, username)
	return err
}

// GetUserByID retrieves a user by their ID
func (db *appdbimpl) GetUserByID(id string) (*User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, username, photo FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func (db *appdbimpl) GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, username, photo FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUsername updates a user's username
func (db *appdbimpl) UpdateUsername(userID, newUsername string) error {
	result, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newUsername, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserPhoto updates a user's profile photo
func (db *appdbimpl) UpdateUserPhoto(userID string, photo []byte) error {
	result, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", photo, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// SearchUsers searches for users by username substring
func (db *appdbimpl) SearchUsers(query string) ([]User, error) {
	rows, err := db.c.Query("SELECT id, username, photo FROM users WHERE username LIKE ? LIMIT 50", "%"+query+"%")
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
