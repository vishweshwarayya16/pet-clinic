package main

import (
	"log"
	"net/http"
	"petclinic/config"
	"petclinic/database"
	"petclinic/handlers"
	"petclinic/middleware"
	"petclinic/utils"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration from .env file
	config.LoadConfig()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	defer database.Close()

	utils.LogMessage(config.LogInfo, "Pet Clinic Management System starting...")

	// Create router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.LoggingMiddleware)

	// Public routes (no authentication required)
	router.HandleFunc("/api/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "Pet Clinic API",
		})
	}).Methods("GET")

	// Protected routes (authentication required)
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// Pet routes
	api.HandleFunc("/pets", handlers.CreatePetHandler).Methods("POST")
	api.HandleFunc("/pets", handlers.GetPetsHandler).Methods("GET")
	api.HandleFunc("/pets/{id}", handlers.GetPetByIDHandler).Methods("GET")
	api.HandleFunc("/pets/{id}", handlers.UpdatePetHandler).Methods("PUT")
	api.HandleFunc("/pets/{id}", handlers.DeletePetHandler).Methods("DELETE")

	// Appointment routes
	api.HandleFunc("/appointments", handlers.CreateAppointmentHandler).Methods("POST")
	api.HandleFunc("/appointments", handlers.GetAppointmentsHandler).Methods("GET")
	api.HandleFunc("/appointments/{id}", handlers.GetAppointmentByIDHandler).Methods("GET")
	api.HandleFunc("/appointments/{id}", handlers.UpdateAppointmentHandler).Methods("PUT")
	api.HandleFunc("/appointments/{id}", handlers.DeleteAppointmentHandler).Methods("DELETE")

	// Medical records routes
	api.HandleFunc("/medical-records", handlers.UploadMedicalRecordHandler).Methods("POST")
	api.HandleFunc("/medical-records/pet/{pet_id}", handlers.GetMedicalRecordsHandler).Methods("GET")
	api.HandleFunc("/medical-records/{id}/download", handlers.DownloadMedicalRecordHandler).Methods("GET")
	api.HandleFunc("/medical-records/{id}", handlers.DeleteMedicalRecordHandler).Methods("DELETE")

	// Start server
	utils.LogMessage(config.LogInfo, "Server listening on "+config.ServerPort)
	if err := http.ListenAndServe(config.ServerPort, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}