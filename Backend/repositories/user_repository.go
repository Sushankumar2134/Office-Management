package repositories

import (
	"database/sql"
	"errors"
	"fmt"

	"Officemanagement/backend/models"
)

var listUsersQuery = `
SELECT id, username, email, is_active, role
FROM users;

`

var userGetByUsernameQuery = `
SELECT u.id, u.username, u.email, u.password_hash, u.is_active, u.role
FROM users u
WHERE u.username = $1;

`

var userGetByIDQuery = `
SELECT u.id, u.username, u.email, u.password_hash, u.is_active, u.role
FROM users u
WHERE u.id = $1;

`

var checkDuplicateUsernameQuery = `
SELECT EXISTS(
    SELECT 1 FROM users WHERE username = $1
);

`

var checkDuplicateEmailQuery = `
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
);

`

var createUserQuery = `
INSERT INTO users (username, email, password_hash, is_active, role)
VALUES ($1, $2, $3, TRUE, $4)
RETURNING id;
`
var getRoleByNameQuery = `
SELECT id FROM roles WHERE role_name = $1;
`

func GetByUsername(db *sql.DB, username string) (*models.User, error) {
	row := db.QueryRow(userGetByUsernameQuery, username)
	//
	user := &models.User{}
	// Scan the row into the user struct
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return user, nil
}

func GetByID(db *sql.DB, id string) (*models.User, error) {
	row := db.QueryRow(userGetByIDQuery, id)

	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return user, nil
}

func CheckDuplicateUsername(db *sql.DB, username string) (bool, error) {
	var exists bool
	err := db.QueryRow(checkDuplicateUsernameQuery, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CheckDuplicateEmail(db *sql.DB, email string) (bool, error) {
	var exists bool
	err := db.QueryRow(checkDuplicateEmailQuery, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateUser(db *sql.DB, username, email, passwordHash, role string) (string, error) {
	var id string
	err := db.QueryRow(createUserQuery, username, email, passwordHash, role).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error creating user: %w", err)
	}
	return id, nil
}

func GetRoleByName(db *sql.DB, roleName string) (string, error) {
	var roleID string
	err := db.QueryRow(getRoleByNameQuery, roleName).Scan(&roleID)
	if err == sql.ErrNoRows {
		return "", errors.New("role not found")
	}
	if err != nil {
		return "", fmt.Errorf("error fetching role: %w", err)
	}
	return roleID, nil
}

func ListUsers(db *sql.DB) ([]models.UserResponse, error) {
	rows, err := db.Query(listUsersQuery)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var u models.UserResponse
		err := rows.Scan(
			&u.UserID,
			&u.Username,
			&u.Email,
			&u.IsActive,
			&u.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}