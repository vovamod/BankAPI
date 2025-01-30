package main

import (
	"fmt"
	"github.com/vovamod/BankAPI/server"
)

func main() {
	app := server.App{}
	err := app.Start()
	if err != nil {
		fmt.Println("Error occurred due to: %w", err)
	}
}
