package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// BookingRequest struct captures booking details from the frontend.
type BookingRequest struct {
	ID             uint    `gorm:"primaryKey"` // Add ID for primary key
	MassagerID     string  `json:"massager_id"`
	ClientName     string  `json:"client_name"`
	ClientEmail    string  `json:"client_email"`
	ClientPhone    string  `json:"client_phone"`
	BookingTime    string  `json:"booking_time"` // ISO 8601 format
	ServiceType    string  `json:"service_type"`
	Notes          string  `json:"notes,omitempty"`  // Optional
	Age            *int    `json:"age,omitempty"`    // Optional
	Gender         *string `json:"gender,omitempty"` // Optional
	RemindByEmail  bool    `json:"remind_by_email"`
	RemindByPhone  bool    `json:"remind_by_phone"`
	DiscountCode   string  `json:"discount_code,omitempty"` // Optional
	OriginalPrice  float64 `json:"original_price"`
	DiscountAmount float64 `json:"discount_amount,omitempty"` // Optional
	FinalPrice     float64 `json:"final_price"`
}

type Repository struct {
	DB *gorm.DB
}

// NewRepository creates a new repository with a given gorm.DB connection.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) SetUpRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/bookings", r.CreateBooking)
}

func main() {

	// Load the .env file
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// Setup database connection with gorm using variables from .env
	dsn := "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("POSTGRES_USER") +
		" dbname=" + os.Getenv("POSTGRES_DB") +
		" sslmode=disable password=" + os.Getenv("POSTGRES_PASSWORD") +
		" port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	// Perform database migration to ensure the tables match our models
	if err := db.AutoMigrate(&BookingRequest{}); err != nil {
		log.Fatal("failed to migrate database", err)
	}

	// Initialize fiber
	app := fiber.New()

	// Setup routes with repository pattern
	repo := NewRepository(db)
	repo.SetUpRoutes(app)

	// Start fiber server
	log.Fatal(app.Listen(":8080"))
}

// CreateBooking handles POST requests to create a new booking.
func (r *Repository) CreateBooking(c *fiber.Ctx) error {
	var booking BookingRequest

	if err := c.BodyParser(&booking); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Call the validation function.
	if err := CreateBookingValidation(&booking); err != nil {
		// If validation fails, return a 422 Unprocessable Entity status, indicating a validation error.
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	if result := r.DB.Create(&booking); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create booking"})
	}

	return c.Status(fiber.StatusCreated).JSON(booking)
}