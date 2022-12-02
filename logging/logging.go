package logging

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger //nolint:gochecknoglobals
	once   sync.Once   //nolint:gochecknoglobals
)

// GetInstance returns the singleton instance of the logger.
func GetInstance() *zap.Logger {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "info"
	}

	once.Do(func() {
		var config zap.Config

		if level == "debug" {
			config = zap.NewDevelopmentConfig()
		} else {
			config = zap.NewProductionConfig()
			config.DisableStacktrace = true
		}

		logger, _ = config.Build()
	})

	defer func() {
		if err := logger.Sync(); err != nil {
			_ = fmt.Errorf("%w", err)
		}
	}()

	return logger
}
