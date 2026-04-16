package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 日志级别环境变量名
const EnvLogLevel = "LOG_LEVEL"

// NewFileLogger 创建文件+标准输出的 Logger
func NewFileLogger(logDir, logFileName string) (Logger, func() error, error) {
	return newFileLoggerWithStdoutWriter(logDir, logFileName, os.Stdout)
}

// NewFileLoggerFromZap 兼容旧代码，返回 (*zap.Logger, cleanup)
func NewFileLoggerFromZap(logDir, logFileName string) (*zap.Logger, func() error, error) {
	l, cleanup, err := newFileLoggerWithStdoutWriter(logDir, logFileName, os.Stdout)
	if err != nil {
		return nil, nil, err
	}
	return l.(*ZapLogger).Unwrap(), cleanup, nil
}

// getLogLevel 从环境变量获取日志级别，默认 Info
func getLogLevel() zapcore.Level {
	levelStr := os.Getenv(EnvLogLevel)
	if levelStr == "" {
		return zap.InfoLevel
	}

	switch strings.ToLower(levelStr) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func newFileLoggerWithStdoutWriter(logDir, logFileName string, stdoutWriter io.Writer) (Logger, func() error, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, nil, err
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	if stdoutWriter == nil {
		stdoutWriter = io.Discard
	}

	level := getLogLevel()

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(logFile),
			zapcore.AddSync(stdoutWriter),
		),
		level,
	)
	z := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	cleanup := func() error {
		if err := z.Sync(); err != nil {
			_ = logFile.Close()
			return err
		}
		return logFile.Close()
	}

	return &ZapLogger{zap: z}, cleanup, nil
}
