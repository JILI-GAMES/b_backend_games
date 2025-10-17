package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// GameLogger represents a game-specific logger
type GameLogger struct {
	gameName string
	logger   *log.Logger
}

// Log levels
const (
	INFO  = "INFO"
	ERROR = "ERROR"
	DEBUG = "DEBUG"
	WARN  = "WARN"
)

var gameLoggers = make(map[string]*GameLogger)

// GetGameLogger returns a logger instance for a specific game
func GetGameLogger(gameName string) *GameLogger {
	// Normalize game name (remove special characters, convert to lowercase)
	normalizedName := strings.ToLower(strings.ReplaceAll(gameName, "_", "-"))

	if logger, exists := gameLoggers[normalizedName]; exists {
		return logger
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return nil
	}

	// Create game-specific log file in logs directory
	logFile := fmt.Sprintf("%s/%s.log", logsDir, normalizedName)

	// Set up lumberjack for log rotation
	lj := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // MB
		MaxBackups: 3,    // Keep 3 backup files
		MaxAge:     7,    // Keep logs for 7 days
		LocalTime:  true, // Use local time for file names
		Compress:   true, // Compress old log files
	}

	// Create logger with game prefix
	gameLogger := log.New(lj, fmt.Sprintf("[%s] ", strings.ToUpper(gameName)), log.LstdFlags|log.Lshortfile)

	// Store in map
	gameLoggers[normalizedName] = &GameLogger{
		gameName: gameName,
		logger:   gameLogger,
	}

	return gameLoggers[normalizedName]
}

// Info logs an info message
func (gl *GameLogger) Info(format string, v ...interface{}) {
	gl.log(INFO, format, v...)
}

// Error logs an error message
func (gl *GameLogger) Error(format string, v ...interface{}) {
	gl.log(ERROR, format, v...)
}

// Debug logs a debug message
func (gl *GameLogger) Debug(format string, v ...interface{}) {
	gl.log(DEBUG, format, v...)
}

// Warn logs a warning message
func (gl *GameLogger) Warn(format string, v ...interface{}) {
	gl.log(WARN, format, v...)
}

// log formats and logs a message with level
func (gl *GameLogger) log(level, format string, v ...interface{}) {
	if gl == nil || gl.logger == nil {
		return
	}

	message := fmt.Sprintf(format, v...)
	gl.logger.Printf("[%s] %s", level, message)
}

// GetGameLoggers returns all active game loggers
func GetGameLoggers() map[string]*GameLogger {
	return gameLoggers
}

// CloseAll closes all game loggers (useful for graceful shutdown)
func CloseAll() {
	for _, logger := range gameLoggers {
		if logger != nil && logger.logger != nil {
			// Lumberjack handles file closing automatically
		}
	}
}
