package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/vovamod/BankAPI/entities"
	"github.com/vovamod/BankAPI/router"
	"github.com/vovamod/BankAPI/utils"
	"os"
)

// App is called from main then Start() and New() will be executed. Separated for easier code reading.
type App struct {
	router *fiber.Router
}

func New(app *fiber.App) *fiber.App {
	log.SetLevel(log.LevelInfo)
	app.Use(logger.New())
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return nil
	}
	// check env. If missing stop process with fatal log
	utils.CheckEnv()
	// Init Mongo Client and pass to others!
	log.Info("Last phase, configuring routes to server")
	loadRoutes(app)
	return app
}

func (a *App) Start() error {
	app := fiber.New(fiber.Config{
		AppName: "BankAPI",
	})
	New(app)
	err := app.Listen(os.Getenv("ADDR"))
	if err != nil {
		log.Fatal("an error occurred in server.go due to: %w", err)
	}
	return nil
}

func loadRoutes(app *fiber.App) *fiber.App {
	entities.Init()
	router.Configure(app)
	return app
}
