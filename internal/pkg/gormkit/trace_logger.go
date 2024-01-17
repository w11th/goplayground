package gormkit

import (
	"context"
	"errors"
	"time"

	kitlog "github.com/theplant/appkit/log"
	"github.com/theplant/appkit/logtracing"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Logger struct {
	logger   kitlog.Logger
	logLevel logger.LogLevel
	logSQL   bool
}

func NewLogger(logger kitlog.Logger, logLevel logger.LogLevel, logSQL bool) logger.Interface {
	return &Logger{
		logger:   logger,
		logLevel: logLevel,
		logSQL:   logSQL,
	}
}

func (l *Logger) LogMode(logLevel logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = logLevel
	return &newLogger
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.logger.Info().Log("msg", msg, "data", data)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.logger.Warn().Log("msg", msg, "data", data)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.logger.Error().Log("msg", msg, "data", data)
	}
}

type spanNameContextKey struct{}

var activeSpanNameContextKey = spanNameContextKey{}

func ContextWithSpanName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, activeSpanNameContextKey, name)
}

func DBWithSpanName(ctx context.Context, db *gorm.DB, name string) *gorm.DB {
	ctx = ContextWithSpanName(ctx, name)
	return db.WithContext(ctx)
}

func spanNameFromContext(ctx context.Context) string {
	name, _ := ctx.Value(activeSpanNameContextKey).(string)
	return name
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	spanName := spanNameFromContext(ctx)
	if spanName == "" {
		return
	}

	if _, ok := kitlog.FromContext(ctx); !ok {
		ctx = kitlog.Context(ctx, l.logger)
	}

	ctx, span := logtracing.StartSpan(ctx, spanName, logtracing.WithStartTime(begin))
	span.AppendKVs(
		"span.type", "sql",
		"span.role", "client",
	)

	sql, rows := fc()
	if l.logSQL {
		span.AppendKVs(
			"sql.sql", sql,
			"sql.rows", rows,
		)
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound)) {
		span.RecordError(err)
	}
	span.End()

	logtracing.LogSpan(ctx, span)
}
