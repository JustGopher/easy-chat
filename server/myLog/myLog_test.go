package myLog

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// 使用已有的配置文件路径
	configPath := "./config_test.ini"

	// 初始化日志
	Init(configPath)

	// 确保日志被正确初始化
	assert.NotNil(t, Logger, "Logger should be initialized")

	// 验证日志级别是否设置为 info
	assert.Equal(t, Logger.GetLevel(), logrus.InfoLevel, "Logger level should be 'info'")

	// 测试日志输出到文件
	fileInfo, err := os.Stat("./server_test.log")
	if err != nil {
		t.Fatalf("failed to stat log file: %v", err)
	}

	// 验证日志文件是否被创建
	assert.True(t, fileInfo.Mode().IsRegular(), "Log file should be created")

	Logger.Info("test")

	// 清理日志文件
	// -- 失败，文件未关闭，log中似乎无需关闭 --
	//err = os.Remove("./server_test.log")
	//if err != nil {
	//	t.Fatalf("failed to remove log file: %v", err)
	//}
}

func TestInvalidConfig(t *testing.T) {
	// 创建一个无效的配置文件路径
	invalidConfigPath := "./server/invalid_config.ini"

	// 初始化日志，应该使用默认级别 info
	Init(invalidConfigPath)

	// 验证日志级别是否设置为 info（默认级别）
	assert.Equal(t, Logger.GetLevel(), logrus.InfoLevel, "Logger level should default to 'info' due to invalid config")
}
