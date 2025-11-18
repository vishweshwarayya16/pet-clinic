package handlers

import (
	//"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"petclinic/config"
	"petclinic/database"
	"petclinic/middleware"
	"petclinic/models"
	"petclinic/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// UploadMedicalRecordHandler handles file uploads for medical records
func UploadMedicalRecordHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(config.MaxUploadSize); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "File too large")
		return
	}

	// Get pet_id from form
	petIDStr := r.FormValue("pet_id")
	petID, err := strconv.Atoi(petIDStr)
	if err != nil || petID == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Valid pet_id is required")
		return
	}

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Check ownership
	if role != "staff" {
		var ownerID int
		err := database.DB.QueryRow("SELECT owner_id FROM pets WHERE id = $1", petID).Scan(&ownerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Pet not found")
			return
		}
		if ownerID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "Access denied")
			return
		}
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(config.UploadDir, os.ModePerm); err != nil {
		utils.LogMessage(config.LogError, "Failed to create upload directory: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d_%s", petID, timestamp, header.Filename)
	filepath := filepath.Join(config.UploadDir, filename)

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		utils.LogMessage(config.LogError, "Failed to create file: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		utils.LogMessage(config.LogError, "Failed to save file: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Save metadata to database
	var id int
	err = database.DB.QueryRow(
		"INSERT INTO medical_records (pet_id, file_name, file_path, file_type) VALUES ($1, $2, $3, $4) RETURNING id",
		petID, header.Filename, filepath, header.Header.Get("Content-Type"),
	).Scan(&id)

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to save record metadata: "+err.Error())
		// Try to delete the uploaded file
		os.Remove(filepath)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save record")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Medical record uploaded: ID=%d, Pet=%d, File=%s", id, petID, header.Filename))
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "File uploaded successfully",
		"id":        id,
		"file_name": header.Filename,
	})
}

// GetMedicalRecordsHandler retrieves all medical records for a pet
func GetMedicalRecordsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petID, _ := strconv.Atoi(vars["pet_id"])

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Check ownership
	if role != "staff" {
		var ownerID int
		err := database.DB.QueryRow("SELECT owner_id FROM pets WHERE id = $1", petID).Scan(&ownerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Pet not found")
			return
		}
		if ownerID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "Access denied")
			return
		}
	}

	// Fetch records
	rows, err := database.DB.Query(
		"SELECT id, pet_id, file_name, file_path, file_type FROM medical_records WHERE pet_id = $1 ORDER BY uploaded_at DESC",
		petID,
	)
	if err != nil {
		utils.LogMessage(config.LogError, "Failed to fetch records: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch records")
		return
	}
	defer rows.Close()

	records := []models.MedicalRecord{}
	for rows.Next() {
		var record models.MedicalRecord
		if err := rows.Scan(&record.ID, &record.PetID, &record.FileName, &record.FilePath, &record.FileType); err != nil {
			continue
		}
		records = append(records, record)
	}

	utils.RespondWithJSON(w, http.StatusOK, records)
}

// DownloadMedicalRecordHandler handles file downloads
func DownloadMedicalRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordID, _ := strconv.Atoi(vars["id"])

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Fetch record and check ownership
	var filepath, filename string
	var petID, ownerID int

	err := database.DB.QueryRow(`
		SELECT mr.file_path, mr.file_name, mr.pet_id, p.owner_id
		FROM medical_records mr
		JOIN pets p ON mr.pet_id = p.id
		WHERE mr.id = $1
	`, recordID).Scan(&filepath, &filename, &petID, &ownerID)

	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Record not found")
		return
	}

	// Check ownership
	if role != "staff" && ownerID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		utils.LogMessage(config.LogError, "File not found on disk: "+filepath)
		utils.RespondWithError(w, http.StatusNotFound, "File not found")
		return
	}

	// Set headers and serve file
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Medical record downloaded: ID=%d, User=%d", recordID, userID))
	http.ServeFile(w, r, filepath)
}

// DeleteMedicalRecordHandler deletes a medical record
func DeleteMedicalRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordID, _ := strconv.Atoi(vars["id"])

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// Fetch record and check ownership
	var filepath string
	var ownerID int

	err := database.DB.QueryRow(`
		SELECT mr.file_path, p.owner_id
		FROM medical_records mr
		JOIN pets p ON mr.pet_id = p.id
		WHERE mr.id = $1
	`, recordID).Scan(&filepath, &ownerID)

	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Record not found")
		return
	}

	// Check ownership
	if role != "staff" && ownerID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Delete from database
	result, err := database.DB.Exec("DELETE FROM medical_records WHERE id = $1", recordID)
	if err != nil {
		utils.LogMessage(config.LogError, "Failed to delete record: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete record")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Record not found")
		return
	}

	// Delete file from disk
	if err := os.Remove(filepath); err != nil {
		utils.LogMessage(config.LogWarn, "Failed to delete file from disk: "+err.Error())
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Medical record deleted: ID=%d", recordID))
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Medical record deleted successfully"})
}
