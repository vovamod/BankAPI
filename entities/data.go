package entities

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// Create a class like obj
type transaction struct {
	Id     primitive.ObjectID `bson:"_id"`
	Value  int                `bson:"value"`
	NameTZ string             `bson:"nameTZ"`
	Date   time.Time          `bson:"date"`
	Status string             `bson:"status"`
	ByWho  string             `bson:"byWho"`
	ToWho  string             `bson:"toWho"`
}

type account struct {
	Id        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Value     int                `bson:"value"`
	AccountId int                `bson:"id"`
}

type user struct {
	Id       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"name"`
	Account  []string           `bson:"account"`
	ObjectId string             `bson:"objectId"`
}

// Init all operations and add them to main App via pointer
func Init(app *fiber.App, db *mongo.Database) *fiber.App {
	InitTransactionRouter(app, db)

	return app
}

// Creates an Group for structs and CRUD ops
func InitTransactionRouter(app *fiber.App, db *mongo.Database) {
	api := app.Group("/api/transactions")
	api.Post("/create", withCollection(db, "transactions", CreateTransaction))
	api.Get("/", withCollection(db, "transactions", GetAllTransactions))
	api.Get("/:id", withCollection(db, "transactions", GetTransactionByID))
}
func InitUserRouter(app *fiber.App, db *mongo.Database) {
	//api := app.Group("/api/user")
	//api.Post("/create", withCollection(db, "user", CreateUser))
	//api.Get("/", withCollection(db, "user", GetAllUsers))
	//api.Get("/:id", withCollection(db, "user", GetUserByID))
	//api.Delete("/:id", withCollection(db, "user", DeleteUserByID))
	//api.Update("/:id", withCollection(db, "user", UpdateUserByID))
}
func InitAccountRouter(app *fiber.App, db *mongo.Database) {
	//api := app.Group("/api/account")
	//api.Post("/create", withCollection(db, "account", CreateAccount))
	//api.Get("/", withCollection(db, "account", GetAllAccount))
	//api.Get("/:id", withCollection(db, "account", GetAccountByID))
	//api.Delete("/:id", withCollection(db, "account", DeleteAccountByID))
	//api.Update("/:id", withCollection(db, "account", UpdateAccountByID))
}

// Middleware
func withCollection(collection *mongo.Database, collectionName string, handler func(*fiber.Ctx, *mongo.Collection) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return handler(c, collection.Collection(collectionName))
	}
}

// CRUD ops
func CreateTransaction(c *fiber.Ctx, collection *mongo.Collection) error {
	var t transaction
	if err := c.BodyParser(&t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validation
	if t.NameTZ == "" || t.ByWho == "" || t.ToWho == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid transaction fields"})
	}
	t.Date = time.Now()
	var a account
	errA := collection.Database().Collection("accounts").FindOne(context.Background(), bson.M{"name": t.ToWho}).Decode(&a)
	if errA != nil {
		t.Status = "Fail"
		_, _ = collection.InsertOne(context.Background(), t.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction. Account does not exist"})
	}

	// Actual logic here thou\
	// If negative value shows up
	if t.Value < 0 && a.Value >= t.Value {
		result, _ := collection.InsertOne(context.Background(), t)
		_, _ = collection.Database().Collection("accounts").UpdateOne(context.Background(), bson.M{"name": t.ToWho}, bson.M{"$set": a.Value - t.Value})
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": "Account has been charged", "data": result})
	}
	// If positive value shows up
	_, _ = collection.Database().Collection("accounts").UpdateOne(context.Background(), bson.M{"name": t.ToWho}, bson.M{"$set": a.Value + t.Value})
	result, _ := collection.InsertOne(context.Background(), t)
	return c.Status(201).JSON(fiber.Map{"data": result})
}
func GetAllTransactions(c *fiber.Ctx, collection *mongo.Collection) error {
	var transactions []transaction

	// Use Find to get all documents in the collection
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Info("Error getting all transactions from database, Error: " + err.Error())
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}
	// Close cursor
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Info("Error closing Mongo cursor, Error: " + err.Error())
		}
	}(cursor, context.Background())
	// Iterate through the cursor and decode each document
	for cursor.Next(context.Background()) {
		var transaction transaction
		if err := cursor.Decode(&transaction); err != nil {
			log.Info("Error getting all transactions from database, Error: " + err.Error())
			return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
		}
		transactions = append(transactions, transaction)
	}

	// Check if the cursor encountered any errors
	if err := cursor.Err(); err != nil {
		log.Info("Error getting all transactions from database, Error: " + err.Error())
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}

	return c.Status(200).JSON(transactions)
}
func GetTransactionByID(c *fiber.Ctx, collection *mongo.Collection) error {
	var transaction transaction
	id := c.Params("id")
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&transaction)
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid transaction ID provided"})
	}
	return c.Status(200).JSON(fiber.Map{"data": transaction})
}
