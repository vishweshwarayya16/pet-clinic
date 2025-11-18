package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"petclinic/config"
	"petclinic/database"
	"petclinic/middleware"
	"petclinic/models"
	"petclinic/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateAppointmentHandler handles creating a new appointment
func CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var appointment models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate required fields
	if appointment.PetID == 0 || appointment.Date.IsZero() {
		utils.RespondWithError(w, http.StatusBadRequest, "Pet ID and date are required")
		return
	}

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Check if user owns the pet (unless staff)
	if role != "staff" {
		var ownerID int
		err := database.DB.QueryRow("SELECT owner_id FROM pets WHERE id = $1", appointment.PetID).Scan(&ownerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Pet not found")
			return
		}
		if ownerID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "Access denied - you don't own this pet")
			return
		}
	}

	// Default status
	if appointment.Status == "" {
		appointment.Status = "scheduled"
	}

	// Insert appointment
	var id int
	err := database.DB.QueryRow(
		"INSERT INTO appointments (pet_id, date, reason, status) VALUES ($1, $2, $3, $4) RETURNING id",
		appointment.PetID, appointment.Date, appointment.Reason, appointment.Status,
	).Scan(&id)

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to create appointment: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create appointment")
		return
	}

	appointment.ID = id
	utils.LogMessage(config.LogInfo, fmt.Sprintf("Appointment created: ID=%d, Pet=%d", id, appointment.PetID))
	utils.RespondWithJSON(w, http.StatusCreated, appointment)
}

// GetAppointmentsHandler retrieves all appointments (filtered by ownership for non-staff)
func GetAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	var rows *sql.Rows
	var err error

	if role == "staff" {
		// Staff can see all appointments
		rows, err = database.DB.Query(`
			SELECT id, pet_id, date, reason, status 
			FROM appointments 
			ORDER BY date DESC
		`)
	} else {
		// Owners can only see appointments for their pets
		rows, err = database.DB.Query(`
			SELECT a.id, a.pet_id, a.date, a.reason, a.status 
			FROM appointments a 
			JOIN pets p ON a.pet_id = p.id 
			WHERE p.owner_id = $1 
			ORDER BY a.date DESC
		`, userID)
	}

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to fetch appointments: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch appointments")
		return
	}
	defer rows.Close()

	appointments := []models.Appointment{}
	for rows.Next() {
		var apt models.Appointment
		if err := rows.Scan(&apt.ID, &apt.PetID, &apt.Date, &apt.Reason, &apt.Status); err != nil {
			continue
		}
		appointments = append(appointments, apt)
	}

	utils.RespondWithJSON(w, http.StatusOK, appointments)
}

// GetAppointmentByIDHandler retrieves a specific appointment
func GetAppointmentByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aptID, _ := strconv.Atoi(vars["id"])

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	var appointment models.Appointment
	var ownerID int

	err := database.DB.QueryRow(`
		SELECT a.id, a.pet_id, a.date, a.reason, a.status, p.owner_id
		FROM appointments a
		JOIN pets p ON a.pet_id = p.id
		WHERE a.id = $1
	`, aptID).Scan(&appointment.ID, &appointment.PetID, &appointment.Date, &appointment.Reason, &appointment.Status, &ownerID)

	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Appointment not found")
		return
	}

	// Check ownership
	if role != "staff" && ownerID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, appointment)
}

// UpdateAppointmentHandler updates an existing appointment
func UpdateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aptID, _ := strconv.Atoi(vars["id"])

	var appointment models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Check ownership
	if role != "staff" {
		var ownerID int
		err := database.DB.QueryRow(`
			SELECT p.owner_id 
			FROM appointments a 
			JOIN pets p ON a.pet_id = p.id 
			WHERE a.id = $1
		`, aptID).Scan(&ownerID)

		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Appointment not found")
			return
		}
		if ownerID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "Access denied")
			return
		}
	}

	// Update appointment
	result, err := database.DB.Exec(
		"UPDATE appointments SET date=$1, reason=$2, status=$3 WHERE id=$4",
		appointment.Date, appointment.Reason, appointment.Status, aptID,
	)

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to update appointment: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update appointment")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Appointment not found")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Appointment updated: ID=%d", aptID))
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Appointment updated successfully"})
}

// DeleteAppointmentHandler deletes an appointment
func DeleteAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aptID, _ := strconv.Atoi(vars["id"])

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Check ownership
	if role != "staff" {
		var ownerID int
		err := database.DB.QueryRow(`
			SELECT p.owner_id 
			FROM appointments a 
			JOIN pets p ON a.pet_id = p.id 
			WHERE a.id = $1
		`, aptID).Scan(&ownerID)

		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Appointment not found")
			return
		}
		if ownerID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "Access denied")
			return
		}
	}

	// Delete appointment
	result, err := database.DB.Exec("DELETE FROM appointments WHERE id = $1", aptID)
	if err != nil {
		utils.LogMessage(config.LogError, "Failed to delete appointment: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete appointment")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Appointment not found")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Appointment deleted: ID=%d", aptID))
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Appointment deleted successfully"})
}