package main

import (
	"bufio"
	"easy-chat/proto"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	connections       = make(map[net.Conn]string) // 用于存储连接
	connHeart         = make(map[net.Conn]time.Time)
	mu                sync.Mutex         // 保护连接的读写
	wt                sync.Mutex         // 确保连续输出的原子性
	heartbeatInterval = 20 * time.Second // 心跳包检测间隔时间
	timeoutInterval   = 90 * time.Second // 连接超时时间
)

func main() {
	//起始界面
	homeText()

	//开始监听
	listen, err := net.Listen("tcp", "localhost:8088")
	if err != nil {
		fmt.Println("listen err=,", err)
		return
	}
	defer listen.Close() //延时关闭

	//循环等待客户端的连接
	fmt.Println("等待客户端连接")
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Accept() err=", err)
			continue
		}

		//起一个协程
		go process(conn)

		wt.Lock()
		println("---------------------------------------------------")
		println("有客户端连接,客户端地址:", conn.RemoteAddr().String())
		//当前所有用户
		fmt.Println("当前用户列表：")
		for n, v := range connections {
			fmt.Printf("%v %v\n", n.RemoteAddr().String(), v)
		}
		println("---------------------------------------------------")
		wt.Unlock()

	}

	//fmt.Printf("listen suc=%v", listen)
}

func process(conn net.Conn) {
	defer conn.Close()
	defer func() {
		//移除连接
		mu.Lock()
		delete(connections, conn)
		delete(connHeart, conn)
		mu.Unlock()

		wt.Lock()
		println("---------------------------------------------------")
		fmt.Println("当前用户列表：")
		for n, v := range connections {
			fmt.Printf("%v %v\n", n.RemoteAddr().String(), v)
		}
		println("---------------------------------------------------")
		wt.Unlock()
	}()
	//接收昵称
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read() err 该客户端接收昵称失败：", conn.RemoteAddr().String())
		return
	}

	nickName := string(buf[:n])
	mu.Lock()
	connections[conn] = nickName
	connHeart[conn] = time.Now()
	mu.Unlock()
	//锁的粒度过大会影响性能，将锁的范围限制在最小的必要区域会更好
	//mu.Lock()
	//connections[conn] = string(buf[:n])
	//mu.Unlock()

	//广播欢迎语
	welcome := "Welcome " + connections[conn] + " joined the chat!\n"
	sendMessage(nil, welcome)

	go heartbeatChecker(conn)

	reader := bufio.NewReader(conn)
	//循环接收客户端发送的数据
	for {
		//等待客户端通过conn发送信息
		//如果客户端没有write，那么协程就会阻塞在这里
		//fmt.Printf("服务器在等待客户端%s发送信息\n", conn.RemoteAddr().String())
		//n, err := conn.Read(buf)
		//if err != nil {
		//	fmt.Println("Read() err 该客户端连接已断开", conn.RemoteAddr().String())
		//	return
		//}
		//显示客户端发送的消息到终端,此处无需加锁
		//message := string(buf[:n])
		message, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("decode msg failed, err:", err)
			return
		}
		if message == "###PING" {
			mu.Lock()
			connHeart[conn] = time.Now() // 更新最后心跳时间
			mu.Unlock()
		} else {
			fmt.Print(message)
			sendMessage(conn, message)
		}

	}
}

// 心跳检测
func heartbeatChecker(conn net.Conn) {
	defer conn.Close()
	defer func() {
		mu.Lock()
		delete(connections, conn)
		delete(connHeart, conn)
		mu.Unlock()
	}()
	for {
		time.Sleep(heartbeatInterval)
		mu.Lock()
		_, exists := connHeart[conn]
		mu.Unlock()
		if !exists {
			return // 如果连接已经被删除，退出
		}
		if time.Since(connHeart[conn]) > timeoutInterval {
			fmt.Println("客户端超时未发送心跳包，断开连接:", conn.RemoteAddr().String())
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
	for c := range connections {
		if c != conn {
			data, err := proto.Encode(message)
			if err != nil {
				fmt.Println("encode msg failed, err:", err)
				return
			}
			_, err = c.Write(data)
			if err != nil {
				fmt.Println("sendMessage failed,err = ", err)
				return
			}
		}
	}
}
