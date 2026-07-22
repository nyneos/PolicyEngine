package logger

import (
	"log"
	"os"
)

func Startup(format string, args ...interface{}) {
	log.Printf("[policy-service] "+format, args...)
}

func Info(format string, args ...interface{}) {
	log.Printf("[policy-service] "+format, args...)
}

func Error(format string, args ...interface{}) {
	log.Printf("[policy-service] ERROR "+format, args...)
}

func Enabled() bool {
	return os.Getenv("POLICY_SERVICE_LOG") == "true"
}
