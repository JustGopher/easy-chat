package main

import (
	"fmt"
	"net"
	"sync"
)

var (
	connections = make(map[net.Conn]bool) // 用于存储连接
	mu          sync.Mutex                // 保护连接的读写
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
		mu.Lock()
		connections[conn] = true // 添加新连接
		mu.Unlock()
		//这里准备起一个协程
		go process(conn)
		println("---------------------------------------------------")
		println("有客户端连接,客户端地址:", conn.RemoteAddr().String())
		//当前所有用户
		fmt.Println("当前用户列表：")
		for n, v := range connections {
			fmt.Printf("%v %v\n", n.RemoteAddr().String(), v)
		}
		println("---------------------------------------------------")
	}

	//fmt.Printf("listen suc=%v", listen)
}

func process(conn net.Conn) {
	defer conn.Close()
	defer func() {
		mu.Lock()
		//delete(connections, conn) // 移除连接
		connections[conn] = false
		mu.Unlock()
		println("---------------------------------------------------")
		fmt.Println("当前用户列表：")
		for n, v := range connections {
			fmt.Printf("%v %v\n", n.RemoteAddr().String(), v)
		}
		println("---------------------------------------------------")
	}()
	//循环接收客户端发送的数据
	for {
		buf := make([]byte, 1024)
		//等待客户端通过conn发送信息
		//如果客户端没有write，那么协程就会阻塞在这里
		//fmt.Printf("服务器在等待客户端%s发送信息\n", conn.RemoteAddr().String())
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Read() err 该客户端连接已断开", conn.RemoteAddr().String())
			return
		}
		//显示客户端发送的消息到终端
		//fmt.Print(string(buf[:n]))
		message := string(buf[:n])
		fmt.Print(message)

		// 广播消息
		mu.Lock()
		for c, v := range connections {
			if c != conn && v == true { // 不发送给自己和离线用户
				_, _ = c.Write([]byte(message)) // 发送给其他客户端
			}
		}
		mu.Unlock()
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
