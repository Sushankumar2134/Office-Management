package main

import (
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"Officemanagement/backend/handlers"
	"Officemanagement/backend/repositories"
)

func main() {
	InitDB()

	log.Println("Application started successfully!")

	authHandler := &handlers.AuthHandler{DB: db}

	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	
	//http.HandleFunc("/Post", authHandler.CreatePost)
	http.HandleFunc("GET /users/{id}", authHandler.GetUserByID)
	http.HandleFunc("PUT /users/{id}", authHandler.UpdateUser)
	http.HandleFunc("DELETE /users/{id}", makeAuthMiddleware(makeRBACMiddleware("admin", authHandler.DeleteUser)))

	http.HandleFunc("/users", makeAuthMiddleware(authHandler.GetUsers))

	http.HandleFunc("/admin/dashboard", makeRBACMiddleware("admin", authHandler.GetAdminDashboard))
	http.HandleFunc("/hr/dashboard", makeRBACMiddleware("HR", authHandler.GetHRDashboard))
	http.HandleFunc("/employee/dashboard", makeRBACMiddleware("employee", authHandler.GetEmployeeDashboard))

	log.Println("Server is listening on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get JWT from authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>"
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		claims := &handlers.Claims{}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "JWT secret not configured", http.StatusInternalServerError)
			return
		}

		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !tkn.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Fetch user from database to verify status
		user, err := repositories.GetByID(db, claims.Subject)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		//Check whether the user is active
		if !user.IsActive {
			http.Error(w, "Account is disabled", http.StatusUnauthorized)
			return
		}

		// Set user context for the handler
		r.Header.Set("X-User-Role", user.Role)
		r.Header.Set("X-User-ID", user.ID)

		next.ServeHTTP(w, r)
	}
}

func makeRBACMiddleware(roleName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check role from JWT claims or header
		userRole := r.Header.Get("X-User-Role")
		if userRole == "" {
			http.Error(w, "User role not found", http.StatusUnauthorized)
			return
		}

		if userRole != roleName {
			http.Error(w, "Forbidden: insufficient role", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}