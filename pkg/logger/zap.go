package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

var (
	coloredLevelStrMap = map[zapcore.Level]Color{
		zapcore.DebugLevel:  Green,
		zapcore.InfoLevel:   Blue,
		zapcore.WarnLevel:   Yellow,
		zapcore.ErrorLevel:  Red,
		zapcore.DPanicLevel: Red,
		zapcore.PanicLevel:  Red,
		zapcore.FatalLevel:  Red,
	}
	defaultLevelColor = Red
)

// Will be used as the context key to store zap.Logger
type loggerKeyType string

const LoggerKey loggerKeyType = "request_key"

func (l loggerKeyType) String() string {
	return string(l)
}

// RootLogger for any events without specific context
var rootLogger *zap.Logger

// RootLoggerf will be used only for events that require no context and formatting
var rootLoggerf *zap.SugaredLogger

func init() {
	rootLogger, _ = getLogger("debug", "console", nil)
	rootLoggerf = rootLogger.Sugar()
}

/*
	logger.NewContextWithLogger creates a new zap.Logger with the specified logFields zap.Field

and stores that logger into given ctx context.Context using constant loggerKey
*/
func NewContextWithLogger(ctx context.Context, logFields ...zap.Field) context.Context {
	return context.WithValue(ctx, LoggerKey, GetContextLogger(ctx).With(logFields...))
}

/*
	logger.GetContextLogger fetches zap.Logger inside given ctx context. Context if exists

otherwise returns fallback/root logger
*/
func GetContextLogger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return rootLogger
	}

	// If the logger was not passed into the context, then just return the global one
	if ctxLogger, ok := ctx.Value(LoggerKey).(*zap.Logger); ok {
		return ctxLogger
	} else {
		return rootLogger
	}
}

func GetContextLoggerf(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return rootLoggerf
	}

	// If the logger was not passed into the context, then just return the global one
	if ctxLogger, ok := ctx.Value(LoggerKey).(*zap.SugaredLogger); ok {
		return ctxLogger
	} else {
		return rootLoggerf
	}
}

func getLogger(level, encoding string, ctx context.Context) (*zap.Logger, error) {
	var logConfig zap.Config
	logConfig = zap.NewDevelopmentConfig()

	// Log level
	logLevel := getLogLevel(level)
	logConfig.Level = zap.NewAtomicLevelAt(logLevel)

	// Set configs
	logConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncoderConfig.EncodeLevel = customDevLevelEncoder

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	// Add fields from context if present
	if ctx != nil {
		if logFields, ok := ctx.Value(LoggerKey).([]zap.Field); ok {
			logger = logger.With(logFields...)
		}
	}

	return logger, nil
}

func getLogLevel(logLevel string) zapcore.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func customDevLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	levelColor, ok := coloredLevelStrMap[level]
	if !ok {
		levelColor = defaultLevelColor
	}
	enc.AppendString("[" + levelColor.Add(level.CapitalString()) + "]")
}

func customProdLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func LogFields(metadata map[string]string) []interface{} {
	// Retrieve context values
	traceID := metadata["traceId"]
	requestUrl := metadata["requestUrl"]
	username := metadata["username"]

	// Init zap fields
	fields := []zap.Field{
		zap.String("username", username),
		zap.String("traceId", traceID),
		zap.String("url", requestUrl),
	}

	// Convert fields to interface
	interfaceFields := make([]interface{}, len(fields))
	for i, field := range fields {
		interfaceFields[i] = field
	}

	return interfaceFields
}
