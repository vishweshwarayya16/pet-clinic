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

// CreatePetHandler handles creating a new pet
func CreatePetHandler(w http.ResponseWriter, r *http.Request) {
	var pet models.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate required fields
	if pet.Name == "" || pet.Species == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name and species are required")
		return
	}

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	// If not staff, set owner_id to current user
	if role != "staff" {
		pet.OwnerID = userID
	} else if pet.OwnerID == 0 {
		// Staff must specify owner_id
		utils.RespondWithError(w, http.StatusBadRequest, "Owner ID is required")
		return
	}

	// Insert pet into database
	var id int
	err := database.DB.QueryRow(
		"INSERT INTO pets (name, species, breed, owner_id, medical_history) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		pet.Name, pet.Species, pet.Breed, pet.OwnerID, pet.MedicalHistory,
	).Scan(&id)

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to create pet: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create pet")
		return
	}

	pet.ID = id
	utils.LogMessage(config.LogInfo, fmt.Sprintf("Pet created: ID=%d, Name=%s, Owner=%d", id, pet.Name, pet.OwnerID))
	utils.RespondWithJSON(w, http.StatusCreated, pet)
}

// GetPetsHandler retrieves all pets (filtered by owner for non-staff users)
func GetPetsHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	var rows *sql.Rows
	var err error

	if role == "staff" {
		// Staff can see all pets
		rows, err = database.DB.Query("SELECT id, name, species, breed, owner_id, medical_history FROM pets ORDER BY id")
	} else {
		// Owners can only see their pets
		rows, err = database.DB.Query(
			"SELECT id, name, species, breed, owner_id, medical_history FROM pets WHERE owner_id = $1 ORDER BY id",
			userID,
		)
	}

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to fetch pets: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch pets")
		return
	}
	defer rows.Close()

	pets := []models.Pet{}
	for rows.Next() {
		var pet models.Pet
		if err := rows.Scan(&pet.ID, &pet.Name, &pet.Species, &pet.Breed, &pet.OwnerID, &pet.MedicalHistory); err != nil {
			continue
		}
		pets = append(pets, pet)
	}

	utils.RespondWithJSON(w, http.StatusOK, pets)
}

// GetPetByIDHandler retrieves a specific pet by ID
func GetPetByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petID, _ := strconv.Atoi(vars["id"])

	userID := middleware.GetUserIDFromRequest(r)
	role := middleware.GetUserRoleFromRequest(r)

	var pet models.Pet
	err := database.DB.QueryRow(
		"SELECT id, name, species, breed, owner_id, medical_history FROM pets WHERE id = $1",
		petID,
	).Scan(&pet.ID, &pet.Name, &pet.Species, &pet.Breed, &pet.OwnerID, &pet.MedicalHistory)

	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Pet not found")
		return
	}

	// Check ownership
	if role != "staff" && pet.OwnerID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, pet)
}

// UpdatePetHandler updates an existing pet
func UpdatePetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petID, _ := strconv.Atoi(vars["id"])

	var pet models.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
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

	// Update pet
	result, err := database.DB.Exec(
		"UPDATE pets SET name=$1, species=$2, breed=$3, medical_history=$4 WHERE id=$5",
		pet.Name, pet.Species, pet.Breed, pet.MedicalHistory, petID,
	)

	if err != nil {
		utils.LogMessage(config.LogError, "Failed to update pet: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update pet")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Pet not found")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Pet updated: ID=%d", petID))
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Pet updated successfully"})
}

// DeletePetHandler deletes a pet
func DeletePetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petID, _ := strconv.Atoi(vars["id"])

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

	// Delete pet
	result, err := database.DB.Exec("DELETE FROM pets WHERE id = $1", petID)
	if err != nil {
		utils.LogMessage(config.LogError, "Failed to delete pet: "+err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete pet")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Pet not found")
		return
	}

	utils.LogMessage(config.LogInfo, fmt.Sprintf("Pet deleted: ID=%d", petID))
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Pet deleted successfully"})
}