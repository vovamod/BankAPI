package router

import (
	"BankAPI/entities"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

// App is called from main then Start() and New() will be executed. Seperated for easier code reading.
type App struct {
	router *fiber.Router
}

func New(app *fiber.App) *fiber.App {
	log.SetLevel(log.LevelInfo)
	app.Use(logger.New())
	// Define EP
	app.Use("/api", func(c *fiber.Ctx) error {
		return c.Next()
	})
	// Init Mongo Client and pass to others!
	db := mongoDatabase().Database("testDB")
	loadRoutes(app, db)
	// Called on shut down op

	return app
}

func (a *App) Start() error {
	app := fiber.New()
	New(app)
	err := app.Listen(":3000")
	if err != nil {
		log.Fatal("an error occured in server.go due to: %w", err)
	}
	// boiler?
	return nil
}

func loadRoutes(api *fiber.App, db *mongo.Database) *fiber.App {
	entities.Init(api, db)
	return api
}

// func for mongo?
func mongoDatabase() *mongo.Client {
	// Check ENV for string
	if err := godotenv.Load(); err != nil {
		log.Errorf("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable")
	}
	// Init client, mongo driver Client.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	// Check whether DB is alive and ping still works
	if c := client.Ping(nil, nil); c != nil {
		log.Fatal("Cannot connect to MongoDB. Error: " + c.Error())
	}
	return client
}
