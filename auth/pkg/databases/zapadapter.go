package databases

import (
	"context"
	"fmt"

	sqldblogger "github.com/simukti/sqldb-logger"
	"go.uber.org/zap"
)

type zapAdapter struct {
	logger      *zap.Logger
	contextKeys []string
}

func NewZapAdapter(logger *zap.Logger, contextKeys []string) sqldblogger.Logger {
	return &zapAdapter{logger: logger, contextKeys: contextKeys}
}

func (zp *zapAdapter) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {

	fields := []zap.Field{}

	for _, key := range zp.contextKeys {
		if val := ctx.Value(key); val != nil {
			fields = append(fields, zap.Any(fmt.Sprint(key), val))
		}
	}

	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case sqldblogger.LevelError:
		zp.logger.Error(msg, fields...)
	case sqldblogger.LevelInfo:
		zp.logger.Info(msg, fields...)
	case sqldblogger.LevelDebug:
		zp.logger.Debug(msg, fields...)
	default:
		// trace will use zap debug
		zp.logger.Debug(msg, fields...)
	}
}
