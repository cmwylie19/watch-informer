package logging

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
)

func TestLogger_NewLogger_NoFilePath(t *testing.T) {
	logger, err := NewLogger("")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetLevel(slog.LevelInfo)
	logger.Info("This is an info message")
}

func TestLogger_NewLogger_FileError(t *testing.T) {
	_, err := NewLogger("/invalid_path/test_log.json")
	if err == nil {
		t.Fatalf("Expected an error due to invalid file path, but got none")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	filePath := "test_log.json"
	defer os.Remove(filePath)

	logger, err := NewLogger(filePath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.CloseFile()

	logger.SetLevel(slog.LevelDebug)
	logger.Debug("This is a debug message")
	validateLogFile(t, filePath, `"level":"DEBUG"`)
	clearFile(t, filePath)

	logger.SetLevel(slog.LevelInfo)
	logger.Info("This is an info message")
	validateLogFile(t, filePath, `"level":"INFO"`)
	clearFile(t, filePath)

	logger.SetLevel(slog.LevelWarn)
	logger.Warn("This is a warning message")
	validateLogFile(t, filePath, `"level":"WARN"`)
	clearFile(t, filePath)

	logger.SetLevel(slog.LevelError)
	logger.Error("This is an error message")
	validateLogFile(t, filePath, `"level":"ERROR"`)
}

func TestLogger_Info(t *testing.T) {
	filePath := "test_log.json"
	defer os.Remove(filePath)

	logger, err := NewLogger(filePath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.CloseFile()

	logger.Info("Info log message")
	validateLogFile(t, filePath, `"level":"INFO"`)
	validateLogFile(t, filePath, `"msg":"Info log message"`)
}

func TestLogger_Debug(t *testing.T) {
	filePath := "test_log.json"
	defer os.Remove(filePath)

	logger, err := NewLogger(filePath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.CloseFile()

	logger.SetLevel(slog.LevelDebug)
	logger.Debug("Debug log message")
	validateLogFile(t, filePath, `"level":"DEBUG"`)
	validateLogFile(t, filePath, `"msg":"Debug log message"`)
}

func TestLogger_Warn(t *testing.T) {
	filePath := "test_log.json"
	defer os.Remove(filePath)

	logger, err := NewLogger(filePath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.CloseFile()

	logger.Warn("Warn log message")
	validateLogFile(t, filePath, `"level":"WARN"`)
	validateLogFile(t, filePath, `"msg":"Warn log message"`)
}

func TestLogger_Error(t *testing.T) {
	filePath := "test_log.json"
	defer os.Remove(filePath)

	logger, err := NewLogger(filePath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.CloseFile()

	logger.Error("Error log message")
	validateLogFile(t, filePath, `"level":"ERROR"`)
	validateLogFile(t, filePath, `"msg":"Error log message"`)
}

func TestLogger_SetCustomLogger(t *testing.T) {
	filePath := "test_log.json"
	defer os.Remove(filePath)

	logger, err := NewLogger(filePath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.CloseFile()

	var buf bytes.Buffer
	customLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
	logger.SetCustomLogger(customLogger)

	logger.Info("Custom log message")

	if !contains(buf.Bytes(), `"msg":"Custom log message"`) {
		t.Errorf("Expected buffer to contain 'Custom log message', but got %s", buf.String())
	}
}

func validateLogFile(t *testing.T, filePath, expected string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !contains(content, expected) {
		t.Errorf("Expected log file to contain %s, but got %s", expected, string(content))
	}
}

func contains(content []byte, expected string) bool {
	return bytes.Contains(content, []byte(expected))
}

func clearFile(t *testing.T, filePath string) {
	err := os.Truncate(filePath, 0)
	if err != nil {
		t.Fatalf("Failed to clear log file: %v", err)
	}
}
