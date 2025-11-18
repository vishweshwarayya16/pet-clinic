package utils

import (
	"log"
	"time"
)

// LogMessage logs a message with a specific level
func LogMessage(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s - %s\n", level, timestamp, message)
}
