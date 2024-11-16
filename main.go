package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lckrugel/billy-the-bot/bot"
	"github.com/lckrugel/billy-the-bot/gateway"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file: ", err)
	}

	bot := bot.NewBot()

	err = gateway.ConnectToGateway(bot)
	if err != nil {
		log.Fatal("error connecting to gateway: ", err)
	}
}
