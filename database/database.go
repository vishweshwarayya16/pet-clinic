package database

import (
	"database/sql"
	"petclinic/config"
	"petclinic/utils"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", config.GetDBConnectionString())
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	utils.LogMessage(config.LogInfo, "Database connected successfully")

	// Create tables
	if err := createTables(); err != nil {
		return err
	}

	utils.LogMessage(config.LogInfo, "Database tables created/verified")
	return nil
}

// createTables creates all necessary database tables
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS owners (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		contact VARCHAR(20),
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'owner',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pets (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		species VARCHAR(50) NOT NULL,
		breed VARCHAR(50),
		owner_id INTEGER REFERENCES owners(id) ON DELETE CASCADE,
		medical_history TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS appointments (
		id SERIAL PRIMARY KEY,
		pet_id INTEGER REFERENCES pets(id) ON DELETE CASCADE,
		date TIMESTAMP NOT NULL,
		reason TEXT,
		status VARCHAR(20) DEFAULT 'scheduled',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS medical_records (
		id SERIAL PRIMARY KEY,
		pet_id INTEGER REFERENCES pets(id) ON DELETE CASCADE,
		file_name VARCHAR(255) NOT NULL,
		file_path VARCHAR(500) NOT NULL,
		file_type VARCHAR(50),
		uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(schema)
	return err
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
		utils.LogMessage(config.LogInfo, "Database connection closed")
	}
}