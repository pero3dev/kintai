package logger

import (
	"testing"
)

func TestNewLogger_Development(t *testing.T) {
	logger, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}
}

func TestNewLogger_Production(t *testing.T) {
	logger, err := NewLogger("info", "production")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}
}

func TestNewLogger_AllLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}
	for _, level := range levels {
		logger, err := NewLogger(level, "development")
		if err != nil {
			t.Fatalf("NewLogger(%s) failed: %v", level, err)
		}
		if logger == nil {
			t.Errorf("Expected logger to be non-nil for level %s", level)
		}
	}
}

func TestLogger_Info(t *testing.T) {
	logger, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Info should not panic
	logger.Info("test message", "key", "value")
}

func TestLogger_Error(t *testing.T) {
	logger, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Error should not panic
	logger.Error("test error", "error", "test")
}

func TestLogger_Warn(t *testing.T) {
	logger, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Warn should not panic
	logger.Warn("test warning", "key", "value")
}

func TestLogger_Debug(t *testing.T) {
	logger, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Debug should not panic
	logger.Debug("test debug", "key", "value")
}

func TestLogger_MultipleKeyValues(t *testing.T) {
	logger, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Should handle multiple key-value pairs
	logger.Info("test message",
		"key1", "value1",
		"key2", 123,
		"key3", true,
	)
}
