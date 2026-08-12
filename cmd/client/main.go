package main

import (
	"log"

	"github.com/ASTeterin/gophkeeper/internal/client"
	"github.com/ASTeterin/gophkeeper/internal/config"
)

func main() {
	cfg := config.ParseFlags()
	app := client.New(&cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}
}
