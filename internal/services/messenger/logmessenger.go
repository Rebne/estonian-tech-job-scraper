package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/Rebne/scrapy_project_v2/pkg/notifier"
)

type LogMessenger struct {
	notifier notifier.Notifier
}

func NewLogMessenger(notifier notifier.Notifier) *LogMessenger {
	return &LogMessenger{
		notifier: notifier,
	}
}

func (lm *LogMessenger) Send(logJSON string) error {
	messages, err := parseLogJSONToHTML(logJSON)
	if err != nil {
		return fmt.Errorf("log messenger failed to parse logJSON: %w", err)
	}
	err = lm.notifier.Notify(messages)
	if err != nil {
		return fmt.Errorf("log messenger failed to notify: %w", err)
	}
	return nil
}

func parseLogJSONToHTML(logJSON string) ([]string, error) {
	logs, err := parseNDJSON(logJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logJSON: %w", err)
	}
	result := make([]string, 0)
	for _, log := range logs {
		html, err := formatLogToHTML(log)
		if err != nil {
			return nil, fmt.Errorf("failed to format log to HTML: %w", err)
		}
		result = append(result, html)
	}

	return result, nil
}

type log struct {
	Time   string `json:"time"`
	Level  string `json:"level"`
	Msg    string `json:"msg"`
	Source string `json:"source,omitempty"`
	Err    string `json:"err,omitempty"`
}

const telegramLogTemplate = `{{if .Time}}⏰ Time: {{.Time}}
{{end}}{{if .Level}}📊 Level: {{.Level}}
{{end}}{{if .Msg}}💬 Message: {{.Msg}}
{{end}}{{if .Source}}📍 Source: {{.Source}}
{{end}}{{if .Err}}🚨 Error: {{.Err}}{{end}}`

func formatLogToHTML(log log) (string, error) {
	tmpl, err := template.New("telegram-log").Parse(telegramLogTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse telegram log template: %w", err)
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, log); err != nil {
		return "", fmt.Errorf("failed to render telegram log template: %w", err)
	}

	return b.String(), nil
}

func parseNDJSON(input string) ([]log, error) {
	var logs []log

	decoder := json.NewDecoder(strings.NewReader(input))

	var err error
	for {
		var l log
		err = decoder.Decode(&l)
		if err != nil {
			if err == io.EOF {
				return logs, nil
			}
			break
		}
		l.trimSpace()
		logs = append(logs, l)
	}

	return logs, err
}

func (l *log) trimSpace() {
	l.Time = strings.TrimSpace(l.Time)
	l.Level = strings.TrimSpace(l.Level)
	l.Msg = strings.TrimSpace(l.Msg)
	l.Source = strings.TrimSpace(l.Source)
	l.Err = strings.TrimSpace(l.Err)
}
