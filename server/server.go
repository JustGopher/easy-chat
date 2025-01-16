package main

import (
	"bufio"
	"easy-chat/proto"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

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

var (
	myConn       *MyConn // 用于存储连接
	myListener   MyListener
	console      *LocalMsg
	broadcastMsg *BroadcastMsg
	config       Config
	logger       *logrus.Logger
)

// MyConn 连接列表
type MyConn struct {
	connections map[net.Conn]string
	rw          sync.RWMutex // 保护连接的读写
}

// CreatMyConn 连接列表初始化
func CreatMyConn() *MyConn {
	return &MyConn{
		connections: make(map[net.Conn]string),
		rw:          sync.RWMutex{},
	}
}

// Add 添加客户端连接
func (c *MyConn) Add(conn net.Conn, name string) {
	c.rw.Lock()
	c.connections[conn] = name
	c.rw.Unlock()
}

// Delete 删除客户端连接
func (c *MyConn) Delete(conn net.Conn) {
	c.rw.Lock()
	delete(c.connections, conn)
	c.rw.Unlock()
}

// UserExit 客户端退出
func (c *MyConn) UserExit(conn net.Conn) {
	if c.connections[conn] != "" {
		console.add(c.connections[conn] + "退出聊天室！")
		broadcastMsg.add(c.connections[conn] + "退出聊天室！")
	}
	myConn.Delete(conn)
	myConn.ShowList()
}

// ShowList 用户连接列表
func (c *MyConn) ShowList() {
	var message string
	message = message + "---------------------------------------------------\n当前用户列表：\n"
	for n, v := range c.connections {
		message = message + fmt.Sprintf("%v %v\n", n.RemoteAddr().String(), v)
	}
	message = message + "---------------------------------------------------"
	console.add(message)
}

// isExist 连接是否存在
func (c *MyConn) isExist(conn net.Conn) bool {
	c.rw.RLock()
	_, exists := myConn.connections[conn]
	c.rw.RUnlock()
	return exists
}

// isNameExist 昵称是否已存在
func (c *MyConn) isNameExist(nickName string) bool {
	for _, v := range myConn.connections {
		if v == nickName {
			return true
		}
	}
	return false
}

// MyListener 监听器
type MyListener struct {
	Listener net.Listener
}

// Close 关闭监听
func (m *MyListener) Close() {
	err := m.Listener.Close()
	if err != nil {
		logger.Fatal("close listener err=", err)
	}
}

// startListen 开始监听
func (m *MyListener) startListen(address string) {
	var err error
	m.Listener, err = net.Listen("tcp", address)
	if err != nil {
		console.add("监听失败...")
		logger.Fatal("listen err=", err)
	} else {
		console.add("监听成功...")
	}
}

// LocalMsg 本地消息
type LocalMsg struct {
	msg chan string
	mu  sync.Mutex
}

// createLocalMsg 创建本地消息实例
func createLocalMsg() *LocalMsg {
	return &LocalMsg{
		msg: make(chan string),
	}
}

// add 添加消息
func (l *LocalMsg) add(s string) {
	l.mu.Lock()
	l.msg <- s
	l.mu.Unlock()
}

// out 输出消息
func (l *LocalMsg) out() {
	for s := range l.msg {
		fmt.Print("\033[G\033[K")
		fmt.Println(s)
		fmt.Print("> ")
	}
}

// clearConsole 清空控制台
func (l *LocalMsg) clearConsole() {
	console.add("\033[2J\033[3J") // 清除屏幕
	console.add("\033[H")         // 将光标移动到左上角
}

// homeText 起始界面
func (l *LocalMsg) homeText() {
	l.clearConsole()
	l.add(`╔═══╗─────────────╔═══╗╔╗───────╔╗─────╔═══╗
║╔══╝─────────────║╔═╗║║║──────╔╝╚╗────║╔═╗║
║╚══╗╔══╗╔══╗╔╗─╔╗║║─╚╝║╚═╗╔══╗╚╗╔╝────║║─╚╝╔══╗
║╔══╝║╔╗║║══╣║║─║║║║─╔╗║╔╗║║╔╗║─║║─╔══╗║║╔═╗║╔╗║
║╚══╗║╔╗║╠══║║╚═╝║║╚═╝║║║║║║╔╗║─║╚╗╚══╝║╚╩═║║╚╝║
╚═══╝╚╝╚╝╚══╝╚═╗╔╝╚═══╝╚╝╚╝╚╝╚╝─╚═╝────╚═══╝╚══╝
─────────────╔═╝║─────by:RationalDysaniaer
─────────────╚══╝ 

服务开始监听...
`)
}

// BroadcastMsg 广播消息
type BroadcastMsg struct {
	msg chan string
	mu  sync.Mutex
}

func createBroadcastMsg() *BroadcastMsg {
	return &BroadcastMsg{
		msg: make(chan string),
	}
}

func (bc *BroadcastMsg) add(message string) {
	bc.mu.Lock()
	bc.msg <- message
	bc.mu.Unlock()
}

func (bc *BroadcastMsg) sendMessage() {
	for message := range bc.msg {
		for c := range myConn.connections {
			data, err := proto.Encode(message)
			if err != nil {
				console.add("编码失败...")
				logger.Error("encode msg failed, go: sendMessage(), err=", err)
				return
			}
			_, err = c.Write(data)
			if err != nil {
				console.add("发送消息失败...")
				logger.Error("sendMessage failed, go: sendMessage(), err=", err)
				return
			}
		}
	}
}

func loadConfig(path string) Config {
	load, err := ini.Load(path)
	if err != nil {
		panic("failed to load ini file")
	}
	err = load.MapTo(&config)
	if err != nil {
		panic("failed to map ini file to struct")
	}
	return config
}

func logInit() {
	logger = logrus.New()
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
}

// init 初始化
func init() {
	myConn = CreatMyConn()
	console = createLocalMsg()
	broadcastMsg = createBroadcastMsg()
	config = loadConfig("./server/config.ini")
	logInit()
}

func main() {
	// 本地消息输出
	go console.out()
	go broadcastMsg.sendMessage()

	// 起始界面
	console.homeText()

	// 开始监听
	myListener.startListen(config.App.Host + ":" + config.App.Port)
	defer myListener.Close()
	console.add("监听端口成功，等待客户端连接...")
	logger.Info("app run")

	// 循环等待客户端的连接
	go waitConn()

	// 获取并处理控制台输入
	waitInput()
}

// waitInput 接收终端输入
func waitInput() {
	rd := bufio.NewReader(os.Stdin)
	for {
		//console.addNoline("> ")
		line, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println("readString err=", err)
			continue
		}
		line = strings.Trim(line, " \r\n")

		switch line {
		case "/help":
			console.add("1. /help\t帮助\n" +
				"2. /users\t查看用户列表\n" +
				"3. /exit\t关闭服务端程序")
		case "/users":
			myConn.ShowList()
		case "/exit":
			fmt.Println("退出程序！")
			os.Exit(0)
		default:
			console.add(`无效命令，输入/help获取帮助`)
		}
	}
}

// waitConn 循环接收客户端连接
func waitConn() {
	for {
		conn, err := myListener.Listener.Accept()
		if err != nil {
			console.add("接收客户端连接失败，正在重试..." + err.Error())
			logger.Error("Accept() err=", err)
			continue
		}
		// 接收到连接后，起一个协程
		go process(conn)
		console.add("有客户端连接,客户端地址:" + conn.RemoteAddr().String())
	}
}

// process 处理客户端连接
func process(conn net.Conn) {
	defer conn.Close()
	defer myConn.UserExit(conn)

	reader := bufio.NewReader(conn)
	var nickName string
	for {
		nickName, _ = proto.Decode(reader)
		data, _ := proto.Encode("false")
		flag := myConn.isNameExist(nickName)
		if !flag {
			data, _ = proto.Encode("true")
		}
		_, err := conn.Write(data)
		if err != nil {
			console.add("发送信息失败...")
			logger.Error("sendMessage failed, go:process for1{}, err = ", err)
			return
		}
		if !flag {
			break
		}
	}

	//添加连接
	myConn.Add(conn, nickName)
	console.add("有用户进入聊天室，用户昵称:" + nickName)
	myConn.ShowList()

	//广播欢迎语
	broadcastMsg.add("Welcome " + myConn.connections[conn] + " joined the chat!")

	lastTime := time.Now()
	go heartbeatChecker(conn, &lastTime)

	//循环接收客户端发送的数据
	for {
		message, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			console.add("解码失败...")
			logger.Error("decode msg failed, go:process for2{}, err:", err)
			return
		}
		if message == "###PING" {
			lastTime = time.Now() // 更新最后心跳时间
		} else {
			console.add(message)
			broadcastMsg.add(message)
		}
	}
}

// heartbeatChecker 心跳检测
func heartbeatChecker(conn net.Conn, lastTime *time.Time) {
	defer conn.Close()
	defer myConn.Delete(conn)
	for {
		time.Sleep(time.Duration(config.App.HeartbeatInterval) * time.Second)
		if !myConn.isExist(conn) {
			return // 如果连接已经被删除，退出
		}
		if time.Since(*lastTime) > time.Duration(config.App.TimeoutInterval)*time.Second {
			console.add("客户端超时未发送心跳包，断开连接:" + conn.RemoteAddr().String())
			logger.Error("heart timeOut:", conn.RemoteAddr().String())
			return
		}
	}
}
