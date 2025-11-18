Pet Clinic Management System – Go Backend

A complete backend API for a Pet Clinic Management System, built with Go, PostgreSQL, JWT Authentication, Role-based access, file uploads, and RESTful endpoints.

🚀 Features
✅ Authentication & Authorization

User registration (Owner / Admin / Staff)

Secure login with JWT tokens

Password hashing with bcrypt

Protected routes using middleware

🐶 Pet Management

Add new pets (owner or admin)

Update pet details

Get list of pets

Fetch pets by owner

📅 Appointment Management

Book appointment for pet

Update appointment

List appointments

Cancel appointment

📤 File Uploads

Upload pet images

Stores in uploads/ folder

Validates file size (configurable)

🗄 PostgreSQL Database

Fully relational schema

Uses github.com/lib/pq

Safe DB connection pooling

📁 Project Structure
petclinic/
│── config/
│   └── config.go
│── database/
│   └── database.go
│── handlers/
│   ├── auth_handler.go
│   ├── pet_handler.go
│   ├── appointment_handler.go
│   └── file_handler.go
│── middleware/
│   └── middleware.go
│── models/
│   └── models.go
│── utils/
│   ├── logger.go
│   └── response.go
│── uploads/
│── main.go
│── go.mod
│── go.sum

⚙️ Installation
1️⃣ Clone the repository
git clone https://github.com/vishweshwarayya16/pet-clinic.git
cd pet-clinic

🔐 Environment Variables

Create your own .env file (NOT committed to GitHub):

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=petclinic

JWT_SECRET=your_jwt_secret

UPLOAD_DIR=uploads
MAX_UPLOAD_SIZE=10485760


For contributors, there is a .env.example file included.

🛢 Database Setup

Create PostgreSQL DB:

CREATE DATABASE petclinic;


Update .env with your DB credentials.

▶️ Running the Application
Install dependencies:
go mod tidy

Run the server:
go run main.go


The server runs on:

http://localhost:9090

🛠 API Endpoints
🔐 Auth
Method	Endpoint	Description
POST	/api/register	Register new user
POST	/api/login	Login user & get token
🐶 Pet Routes
Method	Endpoint	Description
POST	/api/pets	Add new pet
GET	/api/pets	Get all pets
GET	/api/pets/{id}	Get pet by ID
PUT	/api/pets/{id}	Update pet
📅 Appointment Routes
Method	Endpoint	Description
POST	/api/appointments	Book appointment
GET	/api/appointments	List appointments
PUT	/api/appointments/{id}	Update appointment
DELETE	/api/appointments/{id}	Cancel appointment
📤 File Uploads
Method	Endpoint	Description
POST	/api/upload	Upload pet image
🧪 Testing Using Postman
Auth Flow:

Register → /api/register

Login → /api/login

Copy the token returned

In Postman → Headers

Authorization: Bearer <token>


Now you can access protected routes.

📦 Technologies Used

Go 1.21

PostgreSQL

Gorilla Mux

JWT (golang-jwt v5)

bcrypt

godotenv

pq driver

🔒 Security Notes

.env SHOULD NOT be pushed to GitHub

Regenerate your JWT_SECRET if it was exposed

Use environment variables in production (Render, Railway, Docker, etc.)

🤝 Contributing

Pull requests are welcome.
Please open an issue to discuss major changes.

📜 License

This project is Open Source, feel free to use and modify.