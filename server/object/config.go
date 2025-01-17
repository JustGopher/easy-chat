package object

type Config struct {
	App struct {
		Host              string `ini:"host"`
		Port              string `ini:"port"`
		HeartbeatInterval int    `ini:"heartbeatInterval"`
		TimeoutInterval   int    `ini:"timeoutInterval"`
	}
	MyLog struct {
		File   string `ini:"file"`
		Level  string `ini:"level"`
		Format string `ini:"format"`
	}
	Redis struct {
		Host string `ini:"host"`
		Port string `ini:"port"`
		Pwd  string `ini:"pwd"`
		Db   int    `ini:"db"`
	}
}
