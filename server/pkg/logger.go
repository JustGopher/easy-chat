package pkg

import (
	"easy-chat/server/object"
	"github.com/sirupsen/logrus"
	"os"
)

// LogInit 日志初始化
func LogInit(config object.Config) *logrus.Logger {
	logger := logrus.New()
	// 设置日志输出到 server.myLog
	file, err := os.OpenFile(config.MyLog.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Out = os.Stdout
		logger.Warnf("Failed to myLog to file, using default stdout, err: %v", err)
	}
	// 设置日志级别
	// 使用 logrus.ParseLevel 可以避免手动映射日志级别
	level, err := logrus.ParseLevel(config.MyLog.Level)
	if err != nil {
		level = logrus.InfoLevel
		logger.Warnf("Invalid myLog level '%s', using default: %s", config.MyLog.Level, level)
	}
	logger.SetLevel(level)
	// 配置日志格式
	switch config.MyLog.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	return logger
}
