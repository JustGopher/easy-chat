package main

import (
	"bufio"
	"context"
	"easy-chat/proto"
	"easy-chat/server/object"
	"easy-chat/server/pkg"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	connList  *pkg.ConnList
	listener  *pkg.MyListener
	console   *pkg.LocalMsg
	broadcast *pkg.BroadcastMsg
	rdb       *pkg.RedisHandler
	logger    *logrus.Logger
	config    object.Config
	ctx       = context.Background()
)

// init 初始化
func init() {
	config = loadConfig("./server/config.ini")
	connList = pkg.CreatConnList()
	listener = pkg.CreateListener()
	console = pkg.CreateLocalMsg()
	broadcast = pkg.CreateBroadcastMsg()
	logger = pkg.LogInit(config)
	rdb = pkg.NewRedisHandler(config)
	// 启动时清理旧 redis 数据
	err := rdb.Clean(ctx)
	if err != nil {
		log.Fatalf("clean redis data faild when start: %v", err)
	}
}

func main() {
	defer func() {
		if err := rdb.Clean(ctx); err != nil {
			logger.Info("clean redis data failed when close: %v", err)
		}
	}()
	// 消息处理
	go console.Out()
	go func() {
		err := broadcast.SendMessage(connList.GetAllConn())
		if err != nil {
			console.Add("广播发生错误")
			logger.Error("广播错误,err:" + err.Error())
		}
	}()
	go msgQueueProcess()
	// 起始界面
	console.HomeText()
	// 开始监听
	err := listener.StartListen(config.App.Host + ":" + config.App.Port)
	if err != nil {
		logger.Error("listen failed ,err=", err.Error())
	}
	defer listener.Close()
	console.Add("监听端口成功，等待客户端连接...")
	logger.Info("app run")
	// 接收连接
	go waitConn()
	// 接收终端输入
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
			console.Add("0. /help\t帮助\n" +
				"1. /users\t查看用户列表\n" +
				"2. /heart\t查看用户最后心跳时间\n" +
				"3. /rank\t查看用户活跃排行榜\n" +
				"4. /exit\t关闭服务端程序")
		case "/users":
			console.Add(connList.GetList())
		case "/heart":
			console.Add(connList.GetLastHeardTime())
		case "/rank":
			rank, err := rdb.ShowRank(ctx)
			if err != nil {
				console.Add(err.Error())
				logger.Error(err.Error())
			} else {
				console.Add(rank)
			}
		case "/exit":
			console.Add("退出程序！")
			os.Exit(0)
		default:
			console.Add(`无效命令，输入/help获取帮助`)
		}
	}
}

// waitConn 循环接收客户端连接
func waitConn() {
	for {
		conn, err := listener.Accept()
		if err != nil {
			console.Add("接收客户端连接失败，正在重试..." + err.Error())
			logger.Error("Accept() err=", err)
			continue
		}
		// 接收到连接后，起一个协程
		go process(conn)
		console.Add("有客户端连接,客户端地址:" + conn.RemoteAddr().String())
	}
}

// process 处理客户端连接
func process(conn net.Conn) {
	defer conn.Close()
	defer func() {
		console.Add(connList.Connections[conn].NickName + "退出聊天室！")
		broadcast.Add(connList.Connections[conn].NickName + "退出聊天室！")
		connList.Delete(conn)
		console.Add(connList.GetList())
	}()

	reader := bufio.NewReader(conn)
	var nickName string
	for {
		nickName, _ = proto.Decode(reader)
		data, _ := proto.Encode("false")
		flag := connList.IsNameExist(nickName)
		if !flag {
			data, _ = proto.Encode("true")
		}
		_, err := conn.Write(data)
		if err != nil {
			console.Add("发送信息失败...")
			logger.Error("sendMessage failed, go:process for1{}, err = ", err)
			return
		}
		if !flag {
			break
		}
	}

	// 添加连接
	connList.Add(conn, nickName)
	console.Add("有用户进入聊天室，用户昵称:" + nickName)
	console.Add(connList.GetList())

	// 添加用户到排行榜
	err := rdb.AddScore(ctx, nickName)
	if err != nil {
		logger.Error("add user to rank failed,err:", err)
	}
	defer rdb.DelUserFromRank(ctx, nickName)

	// 广播欢迎语
	broadcast.Add("Welcome " + connList.Connections[conn].NickName + " joined the chat!")

	// 开启心跳检测
	go heartbeatChecker(conn)

	// 循环接收客户端发送的数据
	for {
		message, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			console.Add("解码失败...")
			logger.Error("decode msg failed, go:process for2{}, err:", err)
			return
		}
		msg := fmt.Sprintf(nickName + "!$|$|$!" + message)
		err = rdb.MsgQueuePush(ctx, msg)
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

// heartbeatChecker 心跳检测
func heartbeatChecker(conn net.Conn) {
	defer conn.Close()
	defer connList.Delete(conn)
	for {
		time.Sleep(time.Duration(config.App.HeartbeatInterval) * time.Second)
		if !connList.IsExist(conn) {
			return // 如果连接已经被删除，则退出
		}
		if time.Since(connList.Connections[conn].LastHeartTime) > time.Duration(config.App.TimeoutInterval)*time.Second {
			console.Add("客户端超时未发送心跳包，断开连接:" + conn.RemoteAddr().String())
			logger.Error("heart timeOut:", conn.RemoteAddr().String())
			return
		}
	}
}

// msgQueueProcess 消息队列中消息处理
func msgQueueProcess() {
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
			conn, err := connList.GetConnByNickName(nickName)
			if err != nil {
				continue
			}
			if message == "###PING" {
				// 更新最后心跳时间
				connList.Connections[conn].LastHeartTime = time.Now()
			} else {
				console.Add(message)
				broadcast.Add(message)
				err = rdb.AddScore(ctx, nickName)
				if err != nil {
					logger.Error("add score failed,err:", err.Error())
				}
			}
		}
	}
}

// loadConfig 加载配置文件
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
