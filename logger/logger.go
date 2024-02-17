package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger

func InitLogger(logFilePath string, level string) {
	// lumberjack.Logger is already safe for concurrent use, so we don't need to
	// lock it.
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(pe)
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	var logLevel zapcore.Level
	switch level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
		zapcore.NewCore(fileEncoder, w, logLevel),
	)

	l := zap.New(core)
	Logger = l.Sugar()
}
