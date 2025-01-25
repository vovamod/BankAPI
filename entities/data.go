package entities

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
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
	AccountId string             `bson:"accountId"`
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
	InitAccountRouter(app, db)
	InitUserRouter(app, db)
	return app
}

// InitTransactionRouter Creates a Group for structs and CRUD ops
func InitTransactionRouter(app *fiber.App, db *mongo.Database) {
	api := app.Group("/api/transactions")
	api.Post("/create", withCollection(db, "transactions", CreateTransaction))
	api.Get("/", withCollection(db, "transactions", GetAllTransactions))
	api.Get("/:id", withCollection(db, "transactions", GetTransactionByID))
}
func InitUserRouter(app *fiber.App, db *mongo.Database) {
	api := app.Group("/api/user")
	api.Post("/create/:account", withCollection(db, "user", CreateUser))
	api.Get("/", withCollection(db, "user", GetAllUsers))
	api.Get("/:id", withCollection(db, "user", GetUserByID))
	api.Delete("/:id", withCollection(db, "user", DeleteUserByID))
	api.Put("/:id", withCollection(db, "user", UpdateUserByID))
}
func InitAccountRouter(app *fiber.App, db *mongo.Database) {
	api := app.Group("/api/account")
	api.Post("/create", withCollection(db, "account", CreateAccount))
	api.Get("/", withCollection(db, "account", GetAllAccount))
	api.Get("/:id", withCollection(db, "account", GetAccountByID))
	api.Delete("/:id", withCollection(db, "account", DeleteAccountByID))
	// Not needed. We don't want users to update accounts
	//api.Put("/:id", withCollection(db, "account", UpdateAccountByID))
}

// Middleware
func withCollection(collection *mongo.Database, collectionName string, handler func(*fiber.Ctx, *mongo.Collection) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return handler(c, collection.Collection(collectionName))
	}
}

// CRUD ops for InitTransactionRouter
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
	var toWho account
	var byWho account
	errA := collection.Database().Collection("accounts").FindOne(context.Background(), bson.M{"name": t.ToWho}).Decode(&toWho)
	errB := collection.Database().Collection("accounts").FindOne(context.Background(), bson.M{"name": t.ToWho}).Decode(&toWho)
	if errA != nil {
		t.Status = "Fail"
		_, _ = collection.InsertOne(context.Background(), t.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction. Account receiver does not exist"})
	}
	if errB != nil {
		t.Status = "Fail"
		_, _ = collection.InsertOne(context.Background(), t.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction. Account sender does not exist"})
	}

	// Actual logic here thou\
	// If negative value shows up
	if t.Value < 0 && toWho.Value >= t.Value {
		result, _ := collection.InsertOne(context.Background(), t)
		_, _ = collection.Database().Collection("accounts").UpdateOne(context.Background(), bson.M{"name": byWho}, bson.M{"$set": byWho.Value + t.Value})
		_, _ = collection.Database().Collection("accounts").UpdateOne(context.Background(), bson.M{"name": t.ToWho}, bson.M{"$set": toWho.Value - t.Value})
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": "Account has been charged", "data": result})
	}
	// If positive value shows up
	if t.Value >= 0 && byWho.Value >= t.Value {
		_, _ = collection.Database().Collection("accounts").UpdateOne(context.Background(), bson.M{"name": t.ToWho}, bson.M{"$set": toWho.Value + t.Value})
		_, _ = collection.Database().Collection("accounts").UpdateOne(context.Background(), bson.M{"name": byWho}, bson.M{"$set": byWho.Value - t.Value})
		result, _ := collection.InsertOne(context.Background(), t)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": "Account has received value", "data": result})
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction. Maybe Account of receiver or sender does not have the value required for transaction"})

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

// CRUD ops for InitAccountRouter
func CreateAccount(c *fiber.Ctx, collection *mongo.Collection) error {
	var a account
	if err := c.BodyParser(&a); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validation
	if a.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing Name of the Account"})
	}
	a.AccountId = uuid.NewString()
	var duplicate account
	err := collection.FindOne(context.Background(), bson.M{"name": a.Name}).Decode(&duplicate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Account name provided. This account name is already taken"})
	}

	// Actual logic here thou
	result, _ := collection.InsertOne(context.Background(), a)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": "Account has been created", "data": result})
}
func GetAllAccount(c *fiber.Ctx, collection *mongo.Collection) error {
	var a []account

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Info("Error getting all account from database, Error: " + err.Error())
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}
	// Close cursor
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Info("Error closing Mongo cursor, Error: " + err.Error())
		}
	}(cursor, context.Background())

	for cursor.Next(context.Background()) {
		var account account
		if err := cursor.Decode(&account); err != nil {
			log.Info("Error getting all accounts from database, Error: " + err.Error())
			return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
		}
		a = append(a, account)
	}

	if err := cursor.Err(); err != nil {
		log.Info("Error getting all accounts from database, Error: " + err.Error())
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}

	return c.Status(200).JSON(a)
}
func GetAccountByID(c *fiber.Ctx, collection *mongo.Collection) error {
	var a account
	id := c.Params("id")
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&a)
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid account ID provided"})
	}
	return c.Status(200).JSON(fiber.Map{"data": a})
}
func DeleteAccountByID(c *fiber.Ctx, collection *mongo.Collection) error {
	id := c.Params("id")

	err := collection.FindOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid account ID provided"})
	}
	_, errD := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if errD != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid account ID or DB failed?"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Account deleted successfully"})
}

//func UpdateAccountByID(c *fiber.Ctx, collection *mongo.Collection) error {
//	accountID := c.Params("id")
//
//	// Validate
//	objID, err := primitive.ObjectIDFromHex(accountID)
//	if err != nil {
//		log.Errorf("Invalid ID format: %v", err)
//		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//			"error": "Invalid account ID",
//		})
//	}
//
//	var updatedData account
//	if err := c.BodyParser(&updatedData); err != nil {
//		log.Errorf("Failed to parse body: %v", err)
//		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//			"error": "Invalid request body",
//		})
//	}
//
//	update := bson.M{
//		"$set": bson.M{
//			"name":      updatedData.Name,
//			"value":     updatedData.Value,
//			"updatedAt": time.Now(),
//		},
//	}
//
//	filter := bson.M{"_id": objID}
//	result, err := collection.UpdateOne(context.Background(), filter, update)
//	if err != nil {
//		log.Errorf("Failed to update account: %v", err)
//		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//			"error": "Failed to update account",
//		})
//	}
//
//	if result.MatchedCount == 0 {
//		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
//			"error": "Account not found",
//		})
//	}
//
//	return c.Status(fiber.StatusOK).JSON(fiber.Map{
//		"message": "Account updated successfully",
//		"updated": updatedData,
//	})
//}

// CRUD ops for InitUserRouter
func CreateUser(c *fiber.Ctx, collection *mongo.Collection) error {
	var u user
	if err := c.BodyParser(&u); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validation
	if u.Name == "" || u.ObjectId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing Name or ObjectId"})
	}
	// add an objectid check since we will pass to it [16]UUID.string obj
	var duplicate user
	err := collection.FindOne(context.Background(), bson.M{"name": u.Name}).Decode(&duplicate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid username provided. This username is already taken"})
	}

	// Actual logic here thou
	result, _ := collection.InsertOne(context.Background(), u)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": "User has been created", "data": result})
}
func GetAllUsers(c *fiber.Ctx, collection *mongo.Collection) error {
	var u []user
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Info("Error getting all users from database, Error: " + err.Error())
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Info("Error closing Mongo cursor, Error: " + err.Error())
		}
	}(cursor, context.Background())

	for cursor.Next(context.Background()) {
		var user user
		if err := cursor.Decode(&user); err != nil {
			log.Info("Error getting all users from database, Error: " + err.Error())
			return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
		}
		u = append(u, user)
	}

	if err := cursor.Err(); err != nil {
		log.Info("Error getting all users from database, Error: " + err.Error())
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}

	return c.Status(200).JSON(u)
}
func GetUserByID(c *fiber.Ctx, collection *mongo.Collection) error {
	var u user
	id := c.Params("id")
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid user ID provided"})
	}
	return c.Status(200).JSON(fiber.Map{"data": u})
}
func DeleteUserByID(c *fiber.Ctx, collection *mongo.Collection) error {
	id := c.Params("id")
	var u user

	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid user ID provided"})
	}
	for _, account := range u.Account {
		if account != "" {
			return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "User has an account ID provided. User will not be deleted unless his account has been deleted"})
		}
	}
	_, errD := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if errD != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid user ID or DB failed?"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted successfully"})
}
func UpdateUserByID(c *fiber.Ctx, collection *mongo.Collection) error {
	var u user
	var uu user
	id := c.Params("id")
	// Validate if the ID is a valid ObjectID
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid user ID provided"})
	}

	// Parse the request body into an account struct
	if err := c.BodyParser(&uu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	update := bson.M{
		"$set": bson.M{
			"account": append(u.Account, uu.Account...),
		},
	}

	filter := bson.M{"_id": u}
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Errorf("Failed to update user: %v", err)
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Failed to update user"})
	}

	// Check if an account was updated
	if result.MatchedCount == 0 {
		// Rare cases
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "User updated successfully",
		"updated": uu,
	})
}
