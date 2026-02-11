package logger

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger_ConfigBuildError(t *testing.T) {
	expectedErr := errors.New("build failed")

	originalBuilder := zapConfigBuilder
	t.Cleanup(func() {
		zapConfigBuilder = originalBuilder
	})

	zapConfigBuilder = func(config zap.Config) (*zap.Logger, error) {
		return nil, expectedErr
	}

	got, err := NewLogger("info", "production")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if got != nil {
		t.Fatalf("expected nil logger on build error, got %#v", got)
	}
}

func TestLogger_Fatal_WithNoopHook(t *testing.T) {
	log, err := NewLogger("debug", "development")
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Fatal is converted to panic so the test process does not exit.
	log.SugaredLogger = log.SugaredLogger.Desugar().
		WithOptions(zap.OnFatal(zapcore.WriteThenPanic)).
		Sugar()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from fatal log")
		}
	}()

	log.Fatal("fatal test message", "key", "value")
}
