package logger

import (
	"bytes"
	"log/slog"
	"os"
)

type BufferedLogger struct {
	buf bytes.Buffer
	*slog.Logger
}

func NewBufferedLogger(level slog.Level) *BufferedLogger {
	bl := &BufferedLogger{}

	bufferHandler := slog.NewJSONHandler(&bl.buf, &slog.HandlerOptions{
		Level: level,
	})

	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	multiHandler := slog.NewMultiHandler(consoleHandler, bufferHandler)

	bl.Logger = slog.New(multiHandler)
	return bl
}

func (bl *BufferedLogger) Read() string {
	return bl.buf.String()
}

func (bl *BufferedLogger) Reset() {
	bl.buf.Reset()
}
