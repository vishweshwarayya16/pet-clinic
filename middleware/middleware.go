package middleware

import (
	"fmt"
	"net/http"
	"petclinic/config"
	"petclinic/models"
	"petclinic/utils"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// LoggingMiddleware logs all incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.LogMessage(config.LogInfo, fmt.Sprintf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT tokens and adds user info to request headers
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.LogMessage(config.LogWarn, "Missing authorization header")
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &models.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			utils.LogMessage(config.LogWarn, "Invalid token: "+err.Error())
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add user info to request headers for downstream handlers
		r.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		r.Header.Set("X-User-Role", claims.Role)
		r.Header.Set("X-User-Email", claims.Email)

		next.ServeHTTP(w, r)
	})
}

// StaffOnlyMiddleware restricts access to staff members only
func StaffOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("X-User-Role")
		if role != "staff" {
			utils.LogMessage(config.LogWarn, "Unauthorized staff access attempt")
			utils.RespondWithError(w, http.StatusForbidden, "Staff access only")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromRequest extracts user ID from request headers
func GetUserIDFromRequest(r *http.Request) int {
	userID, _ := strconv.Atoi(r.Header.Get("X-User-ID"))
	return userID
}

// GetUserRoleFromRequest extracts user role from request headers
func GetUserRoleFromRequest(r *http.Request) string {
	return r.Header.Get("X-User-Role")
}