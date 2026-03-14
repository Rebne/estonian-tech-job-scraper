package app

import (
	"fmt"
	"os"
)

type EnvironmentVariable string

const DatabaseURL EnvironmentVariable = "DATABASE_URL"
const TelegramBotToken EnvironmentVariable = "TELEGRAM_BOT_TOKEN"
const TelegramChatID EnvironmentVariable = "TELEGRAM_CHAT_ID"

type Config struct {
	DatabaseURL      string
	TelegramBotToken string
	TelegramChatID   string
}

func BuildConfig() (Config, error) {
	databaseURL := os.Getenv(string(DatabaseURL))
	telegramBotToken := os.Getenv(string(TelegramBotToken))
	telegramChatID := os.Getenv(string(TelegramChatID))

	var unsetVariables []string
	if databaseURL == "" {
		unsetVariables = append(unsetVariables, databaseURL)
	}
	if telegramBotToken == "" {
		unsetVariables = append(unsetVariables, telegramBotToken)
	}
	if telegramChatID == "" {
		unsetVariables = append(unsetVariables, telegramChatID)
	}

	if len(unsetVariables) > 0 {
		return Config{}, fmt.Errorf("missing environment variables: %v", unsetVariables)
	}
	return Config{
		DatabaseURL:      databaseURL,
		TelegramBotToken: telegramBotToken,
		TelegramChatID:   telegramChatID,
	}, nil
}
