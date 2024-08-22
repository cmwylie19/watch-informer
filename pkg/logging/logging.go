package logging

import (
	"log/slog"
	"os"
)

type LoggerInterface interface {
	Info(msg string)
	Debug(msg string)
	Warn(msg string)
	Error(msg string)
	SetLevel(level slog.Level)
	CloseFile()
}

var _ LoggerInterface = (*Logger)(nil)

type Logger struct {
	logger   *slog.Logger
	logLevel *slog.LevelVar
	file     *os.File
}

func NewLogger(filePath string) (*Logger, error) {
	logLevel := &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo) // Default level is INFO

	var handler slog.Handler
	var file *os.File
	var err error

	if filePath != "" {
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		handler = slog.NewJSONHandler(file, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}

	l := slog.New(handler)

	return &Logger{
		logger:   l,
		logLevel: logLevel,
		file:     file,
	}, nil
}

func (l *Logger) CloseFile() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) SetLevel(level slog.Level) {
	l.logLevel.Set(level)
}

func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *Logger) SetCustomLogger(customLogger *slog.Logger) {
	l.logger = customLogger
}
