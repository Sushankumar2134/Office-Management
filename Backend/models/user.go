package models

type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email,max=100"`
    Password string `json:"password" validate:"required,min=6,max=100"`
    Role     string `json:"role" validate:"required,oneof=admin HR employee"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Message  string `json:"message"`
    Token    string `json:"token"`
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

type UserResponse struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    IsActive bool   `json:"is_active"`
}

type UserListResponse struct {
    Users []UserResponse `json:"users"`
}

type RegisterResponse struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

type ErrUserNotFound struct {
    Message string
}

func (e *ErrUserNotFound) Error() string {
    return e.Message
}

type User struct {
    ID        string `json:"-"`
    UserID    string `json:"user_id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    PasswordHash string `json:"-"`
    Role      string `json:"-"`
    IsActive  bool   `json:"is_active"`
}
