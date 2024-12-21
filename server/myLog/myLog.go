package myLog

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

var (
	Logger *logrus.Logger
	once   sync.Once
)

type MyLog struct {
	File   string `ini:"file"`
	Level  string `ini:"level"`
	Format string `ini:"format"`
}

func Init(configPath string) {
	once.Do(func() {
		// 初始化
		Logger = logrus.New()

		// 读取配置
		myLog, err := loadConfig(configPath)
		if err != nil {
			panic("Failed to load config path, err:" + err.Error())
		}

		// 设置日志输出到 server.myLog
		file, err := os.OpenFile(myLog.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Logger.Out = file
		} else {
			Logger.Out = os.Stdout
			Logger.Warnf("Failed to myLog to file, using default stdout, err: %v", err)
		}

		// 设置日志级别
		// 使用 logrus.ParseLevel 可以避免手动映射日志级别
		level, err := logrus.ParseLevel(myLog.Level)
		if err != nil {
			level = logrus.InfoLevel
			Logger.Warnf("Invalid myLog level '%s', using default: %s", myLog.Level, level)
		}
		Logger.SetLevel(level)

		//switch myLog.Level {
		//case "panic":
		//	Logger.SetLevel(logrus.PanicLevel)
		//case "fatal":
		//	Logger.SetLevel(logrus.FatalLevel)
		//case "error":
		//	Logger.SetLevel(logrus.ErrorLevel)
		//case "warn":
		//	Logger.SetLevel(logrus.WarnLevel)
		//case "info":
		//	Logger.SetLevel(logrus.InfoLevel)
		//case "debug":
		//	Logger.SetLevel(logrus.DebugLevel)
		//case "trace":
		//	Logger.SetLevel(logrus.TraceLevel)
		//default:
		//	defaultLevel := logrus.InfoLevel
		//	Logger.Warnf("Invalid myLog level '%s', using default: %s", myLog.Level, defaultLevel)
		//	Logger.SetLevel(defaultLevel)
		//}

		// 配置日志格式
		switch myLog.Format {
		case "json":
			Logger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
		case "text":
			Logger.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
		default:
			Logger.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
		}
	})
}

func loadConfig(path string) (*MyLog, error) {
	var myLog MyLog
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load ini file: %w", err)
	}
	if err = cfg.Section("MyLog").MapTo(&myLog); err != nil {
		return nil, fmt.Errorf("failed to map ini file to struct: %w", err)
	}
	// 手动映射:
	// fmt.Println("file:", cfg.Section("MyLog").Key("file").String())
	return &myLog, err
}
