package main

import (
	// Importing necessary packages
	"log"
	"time"

	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/paymentsApi/go-fiber-postgres/models"
	"github.com/paymentsApi/go-fiber-postgres/storage"
	"gorm.io/gorm"
)

// Payment structure defining the payment model
type Payment struct {
	CreditorAcc string    `json:"creditorAcc"`
	DebtorAcc   string    `json:"debtorAcc"`
	Currency    string    `json:"currency"`
	Amount      float32   `json:"amount"`
	Date        time.Time `json:"date"`
	IsDeleted   bool      `json:"isDeleted"`
}

// Repository struct for handling database operations
type Repository struct {
	DB *gorm.DB
}

// CreatePayment handles the creation of a new payment
func (r *Repository) CreatePayment(context *fiber.Ctx) error {
	payment := Payment{}

	err := context.BodyParser(&payment)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Request failed"})
		return err
	}

	err = r.DB.Create(&payment).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not create payment"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Payment has been added."})
	return nil

}

// GetPayments handles fetching all payments
func (r *Repository) GetPayments(context *fiber.Ctx) error {
	paymentModels := &[]models.Payments{}

	err := r.DB.Find(paymentModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not get the payments",
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Payments fetched succesfully",
		"data":    paymentModels,
	})
	return nil

}

// DeletePayment handles the deletion of a payment
func (r *Repository) DeletePayment(context *fiber.Ctx) error {
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "Id cannot be empty.",
		})
		return nil
	}

	paymentModel := &models.Payments{}
	err := r.DB.Where("id = ?", id).First(paymentModel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not find the payment",
		})
		return err
	}

	//soft delete(changing isDeleted field true when delete oparation happen)
	paymentModel.IsDeleted = true
	err = r.DB.Save(paymentModel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not update payment",
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Payment marked as deleted.",
	})
	return nil
}

// GetPaymentID handles fetching a single payment by its ID
func (r *Repository) GetPaymentID(context *fiber.Ctx) error {
	id := context.Params("id")
	paymentModel := &models.Payments{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "Id could not be empty",
		})
		return nil
	}

	err := r.DB.Where("id = ?", id).First(paymentModel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not get the payment",
		})
		return nil
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Payment Id fetch successfully",
		"data":    paymentModel,
	})
	return nil

}

// UpdatePayment handles updating payment details
func (r *Repository) UpdatePayment(context *fiber.Ctx) error {
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "ID is required"})
		return nil
	}

	var updateData struct {
		Amount *float32   `json:"amount"`
		Date   *time.Time `json:"date"`
	}
	if err := context.BodyParser(&updateData); err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Failed to parse request"})
		return err
	}

	updateFields := make(map[string]interface{})
	if updateData.Amount != nil {
		updateFields["amount"] = *updateData.Amount
	}
	if updateData.Date != nil {
		updateFields["date"] = *updateData.Date
	}

	if err := r.DB.Model(&models.Payments{}).Where("id = ?", id).Updates(updateFields).Error; err != nil {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "Failed to update payment"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Payment updated successfully"})
	return nil
}

// SetUpRoutes defines all route handlers for the app
func (r *Repository) SetUpRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_payment", r.CreatePayment)
	api.Delete("/delete-payment/:id", r.DeletePayment)
	api.Get("/get-payment/:id", r.GetPaymentID)
	api.Get("/payments", r.GetPayments)
	api.Put("/update-payment/:id", r.UpdatePayment)
}

func main() {
	// Loading environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	// Database configuration
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
	// Setting up database connection
	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("Could not connect to database")
	}
	// Database migration
	err = models.MigratePayments(db)
	if err != nil {
		log.Fatal("Could not migrate.")
	}
	r := Repository{
		DB: db,
	}
	app := fiber.New()
	// Setting up cors policy
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:4200",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept",
	}))

	r.SetUpRoutes(app)
	app.Listen(":8080")

}
