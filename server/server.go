package main

import (
	"bufio"
	"easy-chat/proto"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var (
	myConn  MyConn     // 用于存储连接
	wt      sync.Mutex // 确保连续输出的原子性
	log     = logrus.New()
	msgChan = make(chan string)
)

const (
	port              = "localhost:8088"
	heartbeatInterval = 20 * time.Second // 心跳包检测间隔时间
	timeoutInterval   = 90 * time.Second // 连接超时时间
)

type MyConn struct {
	connections map[net.Conn]string
	rw          sync.RWMutex // 保护连接的读写
}

func CreatMyConn() *MyConn {
	return &MyConn{
		connections: make(map[net.Conn]string),
		rw:          sync.RWMutex{},
	}
}

func (c *MyConn) Add(conn net.Conn, name string) {
	c.rw.Lock()
	c.connections[conn] = name
	c.rw.Unlock()
}

func (c *MyConn) Delete(conn net.Conn) {
	c.rw.Lock()
	delete(c.connections, conn)
	c.rw.Unlock()
}
func (c *MyConn) UserExit(conn net.Conn) {
	if c.connections[conn] != "" {
		fmt.Println(c.connections[conn] + "退出聊天室！")
		sendMessage(conn, c.connections[conn]+"退出聊天室！\n")
	}
	myConn.Delete(conn)
	myConn.ShowList()
}

func (c *MyConn) ShowList() {
	wt.Lock()
	println("---------------------------------------------------")
	fmt.Println("当前用户列表：")
	for n, v := range c.connections {
		fmt.Printf("%v %v\n", n.RemoteAddr().String(), v)
	}
	println("---------------------------------------------------")
	wt.Unlock()
}
func (c *MyConn) isExist(conn net.Conn) bool {
	c.rw.RLock()
	_, exists := myConn.connections[conn]
	c.rw.RUnlock()
	return exists
}
func (c *MyConn) isNameExist(nickName string) bool {
	for _, v := range myConn.connections {
		if v == nickName {
			return true
		}
	}
	return false
}

type MyListener struct {
	Listener net.Listener
}

func (m *MyListener) Close() {
	err := m.Listener.Close()
	if err != nil {
		log.Fatal("close listener err=", err)
	}
}

func (m *MyListener) CreatListener(address string) {
	var err error
	m.Listener, err = net.Listen("tcp", address)
	if err != nil {
		fmt.Println("监听失败...")
		log.Fatal("listen err=", err)
	} else {
		fmt.Println("监听成功...")
	}
}

func init() {
	myConn = *CreatMyConn()
	file, err := os.OpenFile("./server/logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Out = os.Stdin
		log.Warn("Failed to log to file, using default stderr,err=", err)
	}
}

func main() {
	//起始界面
	homeText()

	//开始监听
	var myListener MyListener
	myListener.CreatListener(port)
	defer myListener.Close()
	fmt.Println("监听端口成功，等待客户端连接...")

	//循环等待客户端的连接
	for {
		conn, err := myListener.Listener.Accept()
		if err != nil {
			fmt.Println("接收客户端连接失败，正在重试...")
			log.Error("Accept() err=", err)
			continue
		}
		//接收到连接后，起一个协程
		go process(conn)
		println("有客户端连接,客户端地址:", conn.RemoteAddr().String())
	}
}

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
			fmt.Println("发送信息失败...")
			log.Error("sendMessage failed, go:process for1{}, err = ", err)
			return
		}
		if !flag {
			break
		}
	}

	//添加连接
	myConn.Add(conn, nickName)

	println("有用户进入聊天室，用户昵称:", nickName)
	myConn.ShowList()

	//广播欢迎语
	sendMessage(nil, "Welcome "+myConn.connections[conn]+" joined the chat!\n")

	lastTime := time.Now()
	go heartbeatChecker(conn, &lastTime)

	//循环接收客户端发送的数据
	for {
		message, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("解码失败...")
			log.Error("decode msg failed, go:process for2{}, err:", err)
			return
		}
		if message == "###PING" {
			lastTime = time.Now() // 更新最后心跳时间
		} else {
			fmt.Print(message)
			sendMessage(conn, message)
		}
	}
}

// 心跳检测
func heartbeatChecker(conn net.Conn, lastTime *time.Time) {
	defer conn.Close()
	defer myConn.Delete(conn)
	for {
		time.Sleep(heartbeatInterval)
		if !myConn.isExist(conn) {
			return // 如果连接已经被删除，退出
		}
		if time.Since(*lastTime) > timeoutInterval {
			fmt.Println("客户端超时未发送心跳包，断开连接:", conn.RemoteAddr().String())
			log.Error("heart timeOut:", conn.RemoteAddr().String())
			return
		}
	}
}

func homeText() {
	clearConsole()
	logo()
	fmt.Println()
	fmt.Println("服务开始监听...")
}

func logo() {
	fmt.Printf(`╔═══╗─────────────╔═══╗╔╗───────╔╗─────╔═══╗
║╔══╝─────────────║╔═╗║║║──────╔╝╚╗────║╔═╗║
║╚══╗╔══╗╔══╗╔╗─╔╗║║─╚╝║╚═╗╔══╗╚╗╔╝────║║─╚╝╔══╗
║╔══╝║╔╗║║══╣║║─║║║║─╔╗║╔╗║║╔╗║─║║─╔══╗║║╔═╗║╔╗║
║╚══╗║╔╗║╠══║║╚═╝║║╚═╝║║║║║║╔╗║─║╚╗╚══╝║╚╩═║║╚╝║
╚═══╝╚╝╚╝╚══╝╚═╗╔╝╚═══╝╚╝╚╝╚╝╚╝─╚═╝────╚═══╝╚══╝
─────────────╔═╝║─────by:RationalDysaniaer
─────────────╚══╝ 
`)
}

func clearConsole() {
	fmt.Print("\033[2J\033[3J") // 清除屏幕
	fmt.Print("\033[H")         // 将光标移动到左上角
}

func sendMessage(conn net.Conn, message string) {
	for c := range myConn.connections {
		if c != conn {
			data, err := proto.Encode(message)
			if err != nil {
				fmt.Println("编码失败...")
				log.Error("encode msg failed, go: sendMessage(), err=", err)
				return
			}
			_, err = c.Write(data)
			if err != nil {
				fmt.Println("发送消息失败...")
				log.Error("sendMessage failed, go: sendMessage(), err=", err)
				return
			}
		}
	}
}
