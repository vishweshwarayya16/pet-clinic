package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Owner represents a pet owner or clinic staff member
type Owner struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Email   string `json:"email"`
	Role    string `json:"role"` // "owner" or "staff"
}

// Pet represents a pet in the clinic
type Pet struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Species        string `json:"species"`
	Breed          string `json:"breed"`
	OwnerID        int    `json:"owner_id"`
	MedicalHistory string `json:"medical_history"`
}

// Appointment represents a clinic appointment
type Appointment struct {
	ID     int       `json:"id"`
	PetID  int       `json:"pet_id"`
	Date   time.Time `json:"date"`
	Reason string    `json:"reason"`
	Status string    `json:"status"`
}

// MedicalRecord represents uploaded medical documents
type MedicalRecord struct {
	ID       int    `json:"id"`
	PetID    int    `json:"pet_id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
}

// User represents registration/login data
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Contact  string `json:"contact"`
	Role     string `json:"role"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims represents JWT token claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}