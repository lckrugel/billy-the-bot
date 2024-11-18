package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lckrugel/discord-bot/internal/config"
	"github.com/lckrugel/discord-bot/internal/gateway"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file: ", err)
	}

	cfg := config.LoadConfig()

	bot := gateway.NewClient(cfg)
	bot.Connect()
	if err != nil {
		log.Fatal("error connecting to gateway: ", err)
	}
}
