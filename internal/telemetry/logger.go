package telemetry

import (
	"os"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds the ECS-encoded zap logger writing to stderr.
func NewLogger() *zap.Logger {
	encoder := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoder, zapcore.AddSync(os.Stderr), zap.InfoLevel)
	return zap.New(core, zap.AddCaller())
}
