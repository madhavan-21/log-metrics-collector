package logger_test

import (
	"fmt"
	"testing"
	"time"

	logger "bitbucket.org/minion/metrics-collector/logger" // Update with actual package path
)

func TestLogCollector(t *testing.T) {
	fmt.Println("🚀 Starting LogCollector test...")

	// Initialize LogCollector with test values
	collector := logger.NewLogCollector("localhost:50052", "TestService", 5, 10*time.Second)
	// defer collector.Stop()

	// Collect different log levels
	collector.Debug("This is a debug message: %d", 1)
	collector.Info("This is an info message: %s", "info_data")
	collector.Warn("This is a warning message with float: %.2f", 3.14)
	collector.Error("This is an error message")

	// Simulate batch logging
	for i := 0; i < 6; i++ { // Exceeds flushSize (5), forcing a flush
		collector.Info("Batch log %d", i)
	}

	// Wait for automatic flush
	select {}
}
