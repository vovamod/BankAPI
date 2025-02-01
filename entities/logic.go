package entities

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/vovamod/BankAPI/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
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
type token struct {
	Token string `json:"token"`
}

// Static checkID
var chBank primitive.ObjectID

// Static mongo client
var db *mongo.Database

// Init all operations and add them to main App via pointer
func Init() {
	db = utils.MongoDatabase()
	BankInit(db, "account")
}

// BankInit check BANK_ISSUER in system for proper Bank support at plugin site (secure enough???)
func BankInit(db *mongo.Database, collectionName string) {
	var ac account
	collection := db.Collection(collectionName)
	log.Debugf("Checking if user BANK_ISSUER exists...")
	err := collection.FindOne(context.Background(), bson.M{"name": "BANK_ISSUER"}).Decode(&ac)
	if err != nil {
		log.Info("No BANK_ISSUER exists. Creating a new BANK_ISSUER...")
		res, errB := collection.InsertOne(context.Background(), &account{Name: "BANK_ISSUER"})
		if errB != nil {
			log.Errorf("Error creating BANK_ISSUER: %v", errB)
		}
		log.Infof("Created BANK_ISSUER: %v", res.InsertedID)
	}
	log.Debugf("BANK_ISSUER exists: %v. Passing to var ch_bank an ObjectID", ac)
	chBank = ac.Id
}

// Middleware
func withCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

// CRUD ops for InitTransactionRouter
func CreateTransaction(c *fiber.Ctx) error {
	collection := withCollection("transactions")
	var t transaction
	if err := c.BodyParser(&t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validation
	if t.NameTZ == "" || t.ByWho == "" || t.ToWho == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or invalid transaction fields"})
	}
	t.Date = time.Now()
	t.Id = primitive.NewObjectID()
	var toWho account
	var byWho account
	errA := collection.Database().Collection("accounts").FindOne(context.Background(), bson.M{"name": t.ToWho}).Decode(&toWho)
	errB := collection.Database().Collection("accounts").FindOne(context.Background(), bson.M{"name": t.ToWho}).Decode(&toWho)
	if errA == nil {
		t.Status = "Fail"
		_, _ = collection.InsertOne(context.Background(), t.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction. Account receiver does not exist"})
	}
	if errB == nil {
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
func GetAllTransactions(c *fiber.Ctx) error {
	_, err := GetAll[transaction](c, withCollection("transactions"))
	return err
}
func GetTransactionByID(c *fiber.Ctx) error {
	return GetByID[transaction](c, withCollection("transactions"))
}

// CRUD ops for InitAccountRouter
func CreateAccount(c *fiber.Ctx) error {
	var a account
	collection := withCollection("account")
	if err := c.BodyParser(&a); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validation
	if a.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing Name of the Account"})
	}
	a.AccountId = uuid.NewString()
	a.Id = primitive.NewObjectID()
	var duplicate account
	err := collection.FindOne(context.Background(), bson.M{"name": a.Name}).Decode(&duplicate)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Account name provided. This account name is already taken"})
	}

	// Actual logic here thou
	result, err1 := collection.InsertOne(context.Background(), a)
	log.Info(err1)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": "Account has been created", "id": result.InsertedID})
}
func GetAllAccount(c *fiber.Ctx) error {
	_, err := GetAll[account](c, withCollection("account"))
	return err
}
func GetAccountByID(c *fiber.Ctx) error {
	return GetByID[account](c, withCollection("account"))
}
func DeleteAccountByID(c *fiber.Ctx) error {
	id := c.Params("id")
	collection := withCollection("account")

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

// TODO: Understand why I wrote such a bad code and why I wanted THAT in the first place!
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
func CreateUser(c *fiber.Ctx) error {
	var u user
	var acc account
	collection := withCollection("user")
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
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid username provided. This username is already taken"})
	}
	// check for account
	acID, _ := primitive.ObjectIDFromHex(c.Params("account"))
	errB := db.Collection("account").FindOne(context.Background(), bson.M{"_id": acID}).Decode(&acc)
	if errB == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid account provided. You cannot create a new user with linked account"})
	}

	// Actual logic here thou
	u.Id = primitive.NewObjectID()
	u.Account = append(u.Account, c.Params("account"))
	result, _ := collection.InsertOne(context.Background(), u)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": "User has been created", "id": result.InsertedID})
}
func GetAllUsers(c *fiber.Ctx) error {
	_, err := GetAll[user](c, withCollection("user"))
	return err
}
func GetUserByID(c *fiber.Ctx) error {
	return GetByID[user](c, withCollection("user"))
}
func DeleteUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var u user
	collection := withCollection("user")

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
func UpdateUserByID(c *fiber.Ctx) error {
	var u user
	var uu user
	collection := withCollection("user")
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

// Others
func AuthBank(c *fiber.Ctx) error {
	// Generate JWT for the authenticated user
	var t token
	sToken := os.Getenv("BANK_256_CODE")
	if err := c.BodyParser(&t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid key provided"})
	}
	if sToken != t.Token {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid key provided"})
	}
	token, err := utils.GenerateToken(chBank.String(), "BANK_ISSUER", 72)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": token,
	})
}
func CheckToken(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"auth": "success"})
}

func GetAll[T any](c *fiber.Ctx, collection *mongo.Collection) ([]T, error) {
	var results []T

	// Perform a find operation on the collection
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Info("Error retrieving documents from database, Error: " + err.Error())
		return nil, c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}
	// Ensure the cursor is closed after processing
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Info("Error closing Mongo cursor, Error: " + err.Error())
		}
	}(cursor, context.Background())

	// Decode each document into the result slice
	for cursor.Next(context.Background()) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			log.Info("Error decoding document from database, Error: " + err.Error())
			return nil, c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
		}
		results = append(results, item)
	}

	// Check if there were any errors during cursor iteration
	if err := cursor.Err(); err != nil {
		log.Info("Error iterating over cursor, Error: " + err.Error())
		return nil, c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid data was received from DB, check logs!"})
	}

	// Check if context provided
	if c == nil {
		return results, nil
	}

	// Return the results as JSON
	return nil, c.Status(fiber.StatusOK).JSON(results)
}
func GetByID[T any](c *fiber.Ctx, collection *mongo.Collection) error {
	var result T
	id, _ := primitive.ObjectIDFromHex(c.Params("id"))
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&result)
	if err == nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{"error": "Invalid ID provided"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}
