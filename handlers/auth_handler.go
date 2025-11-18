package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"petclinic/config"
	"petclinic/database"
	"petclinic/models"
	"petclinic/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.LogMessage(config.LogError, "Invalid registration data: "+err.Error())
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate required fields
	if user.Email == "" || user.Password == "" || user.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	// Default role is owner
	if user.Role == "" {
		user.Role = "owner"
	}

	// Validate role
	if user.Role != "owner" && user.Role != "staff" {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid role. Must be 'owner' or 'staff'")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.LogMessage(config.LogError, "Password hashing failed: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Registration failed")
		return
	}

	// Insert user into database
	var id int
	err = database.DB.QueryRow(
		"INSERT INTO owners (name, contact, email, password, role) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Name, user.Contact, user.Email, string(hashedPassword), user.Role,
	).Scan(&id)

	if err != nil {
		utils.LogMessage(config.LogError, "Database insert failed: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Registration failed - email may already exist")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("User registered: %s (%s)", user.Email, user.Role))
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user_id": id,
		"role":    user.Role,
	})
}

// LoginHandler handles user login and JWT token generation
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate credentials
	if credentials.Email == "" || credentials.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Fetch user from database
	var id int
	var hashedPassword, role, name string
	err := database.DB.QueryRow(
		"SELECT id, password, role, name FROM owners WHERE email = $1",
		credentials.Email,
	).Scan(&id, &hashedPassword, &role, &name)

	if err != nil {
		utils.LogMessage(config.LogWarn, "Login failed for: "+credentials.Email)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password)); err != nil {
		utils.LogMessage(config.LogWarn, "Invalid password for: "+credentials.Email)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
		UserID: id,
		Email:  credentials.Email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		utils.LogMessage(config.LogError, "Token generation failed: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Token generation failed")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("User logged in: %s (%s)", credentials.Email, role))
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":   tokenString,
		"role":    role,
		"name":    name,
		"user_id": id,
	})
}
