package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type ModeOption string

const Dev ModeOption = "dev"
const Test ModeOption = "test"
const Prod ModeOption = "prod"

type EnvironmentVariable string

const DatabaseURL EnvironmentVariable = "DATABASE_URL"
const TelegramBotToken EnvironmentVariable = "TELEGRAM_BOT_TOKEN"
const TelegramChatID EnvironmentVariable = "TELEGRAM_CHAT_ID"
const Mode EnvironmentVariable = "MODE"
const ProxyURL EnvironmentVariable = "PROXY_URL"
const ChromeExecutablePath EnvironmentVariable = "CHROME_EXECUTABLE_PATH"
const TelegramLogThreadID EnvironmentVariable = "TELEGRAM_LOG_THREAD_ID"

type Config struct {
	DatabaseURL          string
	TelegramBotToken     string
	TelegramChatID       string
	ProxyURL             string
	ChromeExecutablePath string
	TelegramLogThreadID  string
	Mode                 ModeOption
	Async                bool
}

func BuildConfig(async bool) (Config, error) {
	_ = godotenv.Load()
	var unsetVariables []string

	databaseURL := os.Getenv(string(DatabaseURL))
	telegramBotToken := os.Getenv(string(TelegramBotToken))
	telegramChatID := os.Getenv(string(TelegramChatID))
	proxyURL := strings.TrimSpace(os.Getenv(string(ProxyURL)))
	chromeExecutablePath := strings.TrimSpace(os.Getenv(string(ChromeExecutablePath)))
	telegramLogThreadID := strings.TrimSpace(os.Getenv(string(TelegramLogThreadID)))
	mode, err := StringToModeOption(os.Getenv(string(Mode)))
	if err != nil {
		unsetVariables = append(unsetVariables, string(Mode))
	}

	if databaseURL == "" && !mode.IsDev() {
		unsetVariables = append(unsetVariables, string(DatabaseURL))
	}
	if telegramBotToken == "" && !mode.IsDev() && !mode.IsTest() {
		unsetVariables = append(unsetVariables, string(TelegramBotToken))
	}
	if telegramChatID == "" && !mode.IsDev() && !mode.IsTest() {
		unsetVariables = append(unsetVariables, string(TelegramChatID))
	}

	if len(unsetVariables) > 0 {
		return Config{}, fmt.Errorf("missing environment variables: %v", unsetVariables)
	}
	return Config{
		DatabaseURL:          databaseURL,
		TelegramBotToken:     telegramBotToken,
		TelegramChatID:       telegramChatID,
		ProxyURL:             proxyURL,
		ChromeExecutablePath: chromeExecutablePath,
		TelegramLogThreadID:  telegramLogThreadID,
		Mode:                 mode,
		Async:                async,
	}, nil
}

func StringToModeOption(s string) (ModeOption, error) {
	switch strings.ToLower(s) {
	case "dev":
		return Dev, nil
	case "test":
		return Test, nil
	case "prod":
		return Prod, nil
	}
	return "", fmt.Errorf("invalid mode option %q: valid options are dev, test, prod", s)
}

func (mo *ModeOption) IsProd() bool {
	return *mo == Prod
}

func (mo *ModeOption) IsDev() bool {
	return *mo == Dev
}

func (mo *ModeOption) IsTest() bool {
	return *mo == Test
}
