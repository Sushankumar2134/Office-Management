package handlers

import (
	"Officemanagement/backend/models"
	"Officemanagement/backend/repositories"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	DB *sql.DB
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		http.Error(w, "Username, email, password, and role are required",
			http.StatusBadRequest)
		return
	}

	// Validate role against roles table

	// Validate role against roles table
	var existingRole string

	err = h.DB.QueryRow(
		"SELECT role_name FROM roles WHERE role_name = $1",
		req.Role,
	).Scan(&existingRole)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid role specified", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println("Role validation error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check duplicate username
	exists, err := repositories.CheckDuplicateUsername(h.DB, req.Username)
	if err != nil {
		log.Println("Error checking duplicate username:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Check duplicate email
	exists, err = repositories.CheckDuplicateEmail(h.DB, req.Email)
	if err != nil {
		log.Println("Duplicate email check error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	userID, err := repositories.CreateUser(h.DB, req.Username, req.Email, hashedPassword, req.Role)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	//creates a response struct to send back to the client
	response := models.RegisterResponse{
		UserID:   userID,
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
// func (h *AuthHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	userID := r.Header.Get("X-User-ID") // set by makeAuthMiddleware
// 	if userID == "" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	var req struct {
// 		Title   string `json:"title"`
// 		Content string `json:"content"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}
// 	if req.Title == "" || req.Content == "" {
// 		http.Error(w, "Title and content are required", http.StatusBadRequest)
// 		return
// 	}

// 	var postID int
// 	err := h.DB.QueryRow(
// 		`INSERT INTO posts (user_id, title, content) VALUES ($1, $2, $3) RETURNING id`,
// 		userID, req.Title, req.Content,
// 	).Scan(&postID)
// 	if err != nil {
// 		log.Println("Create post error:", err)
// 		http.Error(w, "Failed to create post", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"message": "Post created successfully",
// 		"post_id": postID,
// 	})
// }
func (h *AuthHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//// Get ID from URL path
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	user, err := repositories.GetByID(h.DB, userID)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := models.UserResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}

	json.NewEncoder(w).Encode(response)
}
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Get ID from URL path
	userID := r.PathValue("id")

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Role == "" {
		http.Error(w, "Username, email, and role are required",
			http.StatusBadRequest)
		return
	}

	// Check whether the role exists
	var existingRole string

	err = h.DB.QueryRow(
		"SELECT role_name FROM roles WHERE role_name = $1",
		req.Role,
	).Scan(&existingRole)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid role specified", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	query := `
		UPDATE users
		SET username = $1,
		    email = $2,
		    is_active = $3,
		    role = $4,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`

	result, err := h.DB.Exec(
		query,
		req.Username,
		req.Email,
		req.IsActive,
		req.Role,
		userID,
	)

	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	// Check if any row was affected,
    // Get number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User updated successfully",
		"user_id": userID,
	})
}
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// admin only - fetch all users
	users, err := repositories.ListUsers(h.DB)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	response := models.UserListResponse{
		Users: users,
	}

	json.NewEncoder(w).Encode(response)
}
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Stop admins from deleting themselves
	if userID == r.Header.Get("X-User-ID") {
		http.Error(w, "You cannot delete your own account", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		log.Println("Delete user error:", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}
func (h *AuthHandler) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the Admin Dashboard",
	})
}

func (h *AuthHandler) GetHRDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the HR Dashboard",
	})
}

func (h *AuthHandler) GetEmployeeDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the Employee Dashboard",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required",
			http.StatusBadRequest)
		return
	}

	// Fetch user from database
	user, err := repositories.GetByUsername(h.DB, req.Username)
	if err != nil {
		http.Error(w, "Invalid username or password",
			http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is disabled",
			http.StatusUnauthorized)
		return
	}

	// Compare password using bcrypt
	err = CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid username or password",
			http.StatusUnauthorized)
		return
	}

	// Fetch role from roles table using the user's role name
	_, err = repositories.GetRoleByName(h.DB, user.Role)
	if err != nil {
		http.Error(w, "Role not found in database",
			http.StatusInternalServerError)
		return
	}
	// Generate JWT token
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "JWT secret not configured",
			http.StatusInternalServerError)
		return
	}

	claims := Claims{
		Username: user.Username,
		Role:     user.Role,

		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID,
			Issuer:  "Office-management-system",
			IssuedAt: jwt.NewNumericDate(
				time.Now(),
			),
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(15 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString([]byte(secret))

	if err != nil {
		http.Error(w, "Token generation failed",
			http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token:    signedToken,
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
