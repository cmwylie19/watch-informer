package logging

import (
	"log/slog"
)

type MockLogger struct {
	Messages []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Info(msg string)           { m.Messages = append(m.Messages, "INFO: "+msg) }
func (m *MockLogger) Debug(msg string)          { m.Messages = append(m.Messages, "DEBUG: "+msg) }
func (m *MockLogger) Warn(msg string)           { m.Messages = append(m.Messages, "WARN: "+msg) }
func (m *MockLogger) Error(msg string)          { m.Messages = append(m.Messages, "ERROR: "+msg) }
func (m *MockLogger) SetLevel(level slog.Level) {}
func (m *MockLogger) CloseFile()                {}

// Ensure that MockLogger implements LoggerInterface.
var _ LoggerInterface = (*MockLogger)(nil)
