package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vovamod/BankAPI/entities"
)

// Configure runs at the beginning to configure all endpoints and their handlers
func Configure(app *fiber.App) *fiber.App {
	auth := app.Group("/auth")
	auth.Post("/call", entities.AuthBank)
	auth.Get("/call", AuthMiddleware("BANK_ISSUER"), entities.CheckToken)

	transaction := app.Group("/api/transactions")
	transaction.Post("/create", AuthMiddleware("BANK_ISSUER"), entities.CreateTransaction)
	transaction.Get("/", AuthMiddleware("BANK_ISSUER"), entities.GetAllTransactions)
	transaction.Get("/:id", AuthMiddleware("BANK_ISSUER"), entities.GetTransactionByID)

	user := app.Group("/api/user")
	user.Post("/create/:account", AuthMiddleware("BANK_ISSUER"), entities.CreateUser)
	user.Get("/", entities.GetAllUsers)
	user.Get("/:id", entities.GetUserByID)
	user.Delete("/:id", AuthMiddleware("BANK_ISSUER"), entities.DeleteUserByID)
	user.Put("/:id", AuthMiddleware("BANK_ISSUER"), entities.UpdateUserByID)

	account := app.Group("/api/account")
	account.Post("/create", AuthMiddleware("BANK_ISSUER"), entities.CreateAccount)
	account.Get("/", AuthMiddleware("BANK_ISSUER"), entities.GetAllAccount)
	account.Get("/:id", AuthMiddleware("BANK_ISSUER"), entities.GetAccountByID)
	account.Delete("/:id", AuthMiddleware("BANK_ISSUER"), entities.DeleteAccountByID)
	// Not needed. We don't want users to update accounts
	//api.Put("/:id", withCollection("account", UpdateAccountByID))
	return app
}
