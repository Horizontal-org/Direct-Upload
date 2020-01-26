package logger

import (
	"go.uber.org/zap"
)

func NewLogger(verbose bool) (*zap.Logger, error) {
	if verbose {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}
