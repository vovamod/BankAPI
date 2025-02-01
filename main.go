package main

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/vovamod/BankAPI/server"
)

func main() {
	app := server.App{}
	err := app.Start()
	if err != nil {
		log.Fatal("Error occurred due to: %w", err)
	}
}
