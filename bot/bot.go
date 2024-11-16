package bot

import (
	"encoding/json"
	"log"
	"os"
)

type Bot struct {
	intents    uint64
	secret_key string
}

func (bot Bot) GetSecretKey() string {
	return bot.secret_key
}

func (bot Bot) GetIntents() uint64 {
	return bot.intents
}

func NewBot() Bot {
	discord_api_key, isSet := os.LookupEnv("DISCORD_API_KEY")
	if !isSet {
		log.Fatal("missing environment variable: 'DISCORD_API_KEY'")
	}

	intentsMap := readBotConfig()
	intents := calculateBotIntents(intentsMap)

	bot := Bot{
		secret_key: discord_api_key,
		intents:    intents,
	}

	return bot
}

func readBotConfig() map[string]bool {
	fileContent, err := os.ReadFile("./config/bot_intents_config.json")
	if err != nil {
		log.Fatal("could not read file 'bot_intents_config.json'")
	}

	var config map[string]bool
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Fatal("error converting config file")
	}

	return config
}

func calculateBotIntents(intentsMap map[string]bool) uint64 {
	// Todas as intents e seus valores
	intentValues := map[string]uint64{
		"guilds":                        1 << 0,
		"guid_members":                  1 << 1,
		"guild_moderation":              1 << 2,
		"guild_expressions":             1 << 3,
		"guild_integrations":            1 << 4,
		"guild_webhooks":                1 << 5,
		"guild_invites":                 1 << 6,
		"guild_voice_states":            1 << 7,
		"guild_presences":               1 << 8,
		"guild_messages":                1 << 9,
		"guild_message_reactions":       1 << 10,
		"guild_message_typing":          1 << 11,
		"direct_messages":               1 << 12,
		"direct_message_reactions":      1 << 13,
		"direct_message_typing":         1 << 14,
		"message_content":               1 << 15,
		"guild_scheduled_events":        1 << 16,
		"auto_moderation_configuration": 1 << 20,
		"auto_moderation_execution":     1 << 21,
		"guild_message_polls":           1 << 24,
		"direct_message_polls":          1 << 25,
	}

	var intents uint64 = 0
	for key, value := range intentsMap {
		if value {
			if bitValue, exists := intentValues[key]; exists {
				intents |= bitValue
			} else {
				log.Print("unkown intent key: ", key)
			}
		}
	}

	return intents
}
