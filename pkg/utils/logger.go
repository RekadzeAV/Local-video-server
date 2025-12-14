package utils

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	// Logger - глобальный экземпляр логгера
	Logger *logrus.Logger
)

// InitLogger инициализирует систему логирования
func InitLogger(level string, format string, logFile string) error {
	Logger = logrus.New()

	// Установка уровня логирования
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Logger.SetLevel(logLevel)

	// Установка формата вывода
	if format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	// Установка вывода
	var output io.Writer = os.Stdout
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}
	Logger.SetOutput(output)

	return nil
}

// GetLogger возвращает глобальный логгер
func GetLogger() *logrus.Logger {
	if Logger == nil {
		InitLogger("info", "text", "")
	}
	return Logger
}
