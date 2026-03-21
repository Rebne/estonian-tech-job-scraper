package app

import (
	"fmt"
	"os"
)

type EnvironmentVariable string

const DatabaseURL EnvironmentVariable = "DATABASE_URL"
const TelegramBotToken EnvironmentVariable = "TELEGRAM_BOT_TOKEN"
const TelegramChatID EnvironmentVariable = "TELEGRAM_CHAT_ID"
const Mode EnvironmentVariable = "MODE"

type Config struct {
	DatabaseURL      string
	TelegramBotToken string
	TelegramChatID   string
	Mode ModeOption
}

func BuildConfig() (Config, error) {
	var unsetVariables []string

	databaseURL := os.Getenv(string(DatabaseURL))
	telegramBotToken := os.Getenv(string(TelegramBotToken))
	telegramChatID := os.Getenv(string(TelegramChatID))
	mode, err := StringToModeOption(os.Getenv(string(Mode)))
	if err != nil {
		unsetVariables = append(unsetVariables, string(mode))
	}

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
		Mode: mode,
	}, nil
}
