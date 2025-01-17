package main

import (
	"bufio"
	"context"
	"easy-chat/proto"
	"easy-chat/server/object"
	"easy-chat/server/redisDB"
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	myConn       *MyConn // 用于存储连接
	myListener   MyListener
	console      *LocalMsg
	broadcastMsg *BroadcastMsg
	config       object.Config
	logger       *logrus.Logger
	rdb          *redisDB.RedisHandler
	ctx          = context.Background()
)

// ConnState 连接状态
type ConnState struct {
	nickName      string
	add           string
	loginTime     time.Time
	lastHeartTime time.Time
}

// MyConn 连接列表
type MyConn struct {
	connections map[net.Conn]*ConnState
	rw          sync.RWMutex // 保护连接的读写
}

// CreatMyConn 连接列表初始化
func CreatMyConn() *MyConn {
	return &MyConn{
		connections: make(map[net.Conn]*ConnState),
		rw:          sync.RWMutex{},
	}
}

// Add 添加客户端连接
func (c *MyConn) Add(conn net.Conn, state *ConnState) {
	c.rw.Lock()
	c.connections[conn] = state
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
	console.add(c.connections[conn].nickName + "退出聊天室！")
	broadcastMsg.add(c.connections[conn].nickName + "退出聊天室！")
	myConn.Delete(conn)
	myConn.ShowList()
}

// ShowList 用户连接列表
func (c *MyConn) ShowList() {
	var message string
	message = message + "---------------------------------------------------\n当前用户列表：\n"
	message = message + fmt.Sprintf("IP              登录时间            昵称\n")
	for n, v := range c.connections {
		message = message + fmt.Sprintf("%v %v %v\n", n.RemoteAddr().String(), v.loginTime.Format("2006:01:02 15:04:05"), v.nickName)
	}
	message = message + "---------------------------------------------------"
	console.add(message)
}

// ShowLastHeardTime 显示心跳时间
func (c *MyConn) ShowLastHeardTime() {
	var message string
	message = message + "---------------------------------------------------\n用户心跳列表：\n"
	message = message + fmt.Sprintf("登录时间            最后心跳时间        昵称\n")
	for _, v := range c.connections {
		message = message + fmt.Sprintf("%v %v %v\n", v.loginTime.Format("2006:01:02 15:04:05"), v.lastHeartTime.Format("2006:01:02 15:04:05"), v.nickName)
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
		if v.nickName == nickName {
			return true
		}
	}
	return false
}

// getConnByNickName 通过昵称获取连接
func (c *MyConn) getConnByNickName(nickName string) (net.Conn, error) {
	for k, v := range myConn.connections {
		if v.nickName == nickName {
			return k, nil
		}
	}
	return nil, errors.New("no user")
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

// createBroadcastMsg 创建广播消息处理
func createBroadcastMsg() *BroadcastMsg {
	return &BroadcastMsg{
		msg: make(chan string),
	}
}

// add 添加广播消息
func (bc *BroadcastMsg) add(message string) {
	bc.mu.Lock()
	bc.msg <- message
	bc.mu.Unlock()
}

// sendMessage 发送广播消息
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

// loadConfig 加载日志文件
func loadConfig(path string) object.Config {
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

// logInit 日志初始化
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

// redisInit redis初始化
func redisInit() {
	rdb = redisDB.NewRedisHandler(config)
	// 启动时清理旧数据
	if err := rdb.Clean(ctx); err != nil {
		log.Fatalf("clean redis data faild when start: %v", err)
	}
	// 服务结束时清理数据
	defer func() {
		if err := rdb.Clean(ctx); err != nil {
			log.Printf("clean redis data faild when close: %v", err)
		}
	}()
}

// init 初始化
func init() {
	myConn = CreatMyConn()
	console = createLocalMsg()
	broadcastMsg = createBroadcastMsg()
	config = loadConfig("./server/config.ini")
	logInit()
	redisInit()
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

	// 开启消息队列消费者
	for i := 0; i < 3; i++ {
		go msgQueue()
	}

	// 循环等待客户端的连接
	go waitConn()

	// 获取并处理控制台输入
	waitInput()
}

// waitInput 接收终端输入
func waitInput() {
	rd := bufio.NewReader(os.Stdin)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println("readString err=", err)
			continue
		}
		line = strings.Trim(line, " \r\n")

		switch line {
		case "/help":
			console.add("0. /help\t帮助\n" +
				"1. /users\t查看用户列表\n" +
				"2. /heart\t查看用户最后心跳时间\n" +
				"3. /rank\t查看用户活跃排行榜\n" +
				"4. /exit\t关闭服务端程序")
		case "/users":
			myConn.ShowList()
		case "/heart":
			myConn.ShowLastHeardTime()
		case "/rank":
			rank, err := rdb.ShowRank(ctx)
			if err != nil {
				console.add(err.Error())
				logger.Error(err.Error())
			} else {
				console.add(rank)
			}
		case "/exit":
			console.add("退出程序！")
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
	state := &ConnState{
		nickName:      nickName,
		add:           conn.RemoteAddr().String(),
		loginTime:     time.Now(),
		lastHeartTime: time.Now(),
	}
	//添加连接
	myConn.Add(conn, state)
	console.add("有用户进入聊天室，用户昵称:" + nickName)
	myConn.ShowList()

	//添加用户到排行榜
	err := rdb.AddScore(ctx, nickName)
	if err != nil {
		logger.Error("add user to rank failed,err:", err)
	}
	defer rdb.DelUserFromRank(ctx, nickName)

	//广播欢迎语
	broadcastMsg.add("Welcome " + myConn.connections[conn].nickName + " joined the chat!")

	// 开启心跳检测
	go heartbeatChecker(conn)

	// 循环接收客户端发送的数据
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
		msg := fmt.Sprintf(state.nickName + "!$|$|$!" + message)
		err = rdb.MsgQueuePush(ctx, msg)
		if err != nil {
			logger.Error(err.Error())
		}
		err = rdb.AddScore(ctx, nickName)
	}
}

// heartbeatChecker 心跳检测
func heartbeatChecker(conn net.Conn) {
	defer conn.Close()
	defer myConn.Delete(conn)
	for {
		time.Sleep(time.Duration(config.App.HeartbeatInterval) * time.Second)
		if !myConn.isExist(conn) {
			return // 如果连接已经被删除，退出
		}
		if time.Since(myConn.connections[conn].lastHeartTime) > time.Duration(config.App.TimeoutInterval)*time.Second {
			console.add("客户端超时未发送心跳包，断开连接:" + conn.RemoteAddr().String())
			logger.Error("heart timeOut:", conn.RemoteAddr().String())
			return
		}
	}
}

// msgQueue 消息队列处理
func msgQueue() {
	for {
		result, err := rdb.MsgQueuePop(ctx)
		if err != nil || len(result) < 2 {
			continue
		}
		fullMsg := result[1]
		parts := strings.SplitN(fullMsg, "!$|$|$!", 2) // 按照 "!$|$|$!" 分割
		if len(parts) == 2 {
			nickName := parts[0]
			message := parts[1]
			conn, err := myConn.getConnByNickName(nickName)
			if err != nil {
				continue
			}
			if message == "###PING" {
				myConn.connections[conn].lastHeartTime = time.Now() // 更新最后心跳时间
			} else {
				console.add(message)
				broadcastMsg.add(message)
			}
		}
	}
}
