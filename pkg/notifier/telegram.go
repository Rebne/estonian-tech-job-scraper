package notifier

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const MAX_LENGTH = 4096

var ErrInvalidHTML = errors.New("invalid HTML")

type telegramNotifier struct {
	notifier     INotifier
	messagesHTML []string
	config       *telegramNotifierConfig
}

func NewTelegramNotifier(notifier INotifier, messagesHTML []string, config *telegramNotifierConfig) (*telegramNotifier, error) {
	for _, message := range messagesHTML {
		if !isParseableHTML(message) {
			return nil, ErrInvalidHTML
		}
	}
	return &telegramNotifier{
		notifier,
		messagesHTML,
		config,
	}, nil
}

type telegramNotifierConfig struct {
	botToken string
	chatID   string
}

func NewTelegramNotifierConfig(botToken, chatID string) *telegramNotifierConfig {
	return &telegramNotifierConfig{
		botToken,
		chatID,
	}
}

func (tn *telegramNotifier) Notify() error {
	defer tn.notifier.Notify()
	if len(tn.messagesHTML) == 0 {
		return nil
	}
	for _, message := range chunk(tn.messagesHTML) {
		if err := tn.sendTelegramMessage(message); err != nil {
			return fmt.Errorf("failed to send telegram message: %w", err)
		}
	}

	return nil
}

func (tn *telegramNotifier) sendTelegramMessage(message string) error {
	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		tn.config.botToken,
	)

	data := url.Values{}
	data.Set("chat_id", tn.config.chatID)
	data.Set("parse_mode", "HTML")
	data.Set("text", message)

	resp, err := http.Post(
		endpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %s", resp.Status)
	}

	return nil
}

func chunk(messages []string) []string {
	if len(messages) == 1 && len(messages[0]) <= MAX_LENGTH || len(messages) == 0 {
		return messages
	}

	result := []string{}
	currentChunk := ""

	for _, message := range messages {

		if currentChunk != "" && len(currentChunk)+len(message)+1 > MAX_LENGTH {
			result = append(result, currentChunk)
			currentChunk = message
		} else {
			if currentChunk != "" {
				currentChunk += "\n" + message
			} else {
				currentChunk = message
			}
		}
	}

	if currentChunk != "" {
		result = append(result, currentChunk)
	}

	return result
}

func isParseableHTML(input string) bool {
	_, err := html.ParseFragment(strings.NewReader(input), nil)
	return err == nil
}
