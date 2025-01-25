package main

import (
	"BankAPI/router"
	"fmt"
)

func main() {
	app := router.App{}
	err := app.Start()
	if err != nil {
		fmt.Println("Error occurred due to: %w", err)
	}
}
