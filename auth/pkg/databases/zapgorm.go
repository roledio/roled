package databases

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

// CreateZapGormLogger creates a new GORM logger using the provided Zap logger and context key.
func CreateZapGormLogger(zaplogger *zap.Logger, contextKey interface{}) logger.Interface {
	zg := zapgorm2.New(zaplogger)
	zg.LogLevel = logger.Info
	zg.IgnoreRecordNotFoundError = true
	zg.SlowThreshold = 1 * time.Second
	if contextKey != nil {
		zg.Context = func(ctx context.Context) []zapcore.Field {
			return []zapcore.Field{zap.Any(fmt.Sprint(contextKey), ctx.Value(contextKey))}
		}
	}
	zg.SetAsDefault()
	return zg
}
