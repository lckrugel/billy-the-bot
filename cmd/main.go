package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	stopChan := make(chan struct{})

	var wg sync.WaitGroup

	bot := gateway.NewClient(cfg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bot.Connect(); err != nil {
			log.Fatalf("Failed to start bot: %v", err)
		}

		// Wait for stop signal.
		<-stopChan
		bot.Disconnect()
		log.Println("Disconnected")
	}()

	// Wait for OS signal.
	<-stop
	log.Println("Received shutdown signal")

	// Signal all components to stop.
	close(stopChan)

	// Wait for all goroutines to complete.
	wg.Wait()
	log.Println("Application stopped cleanly")
}
