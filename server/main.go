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
	fmt.Println("服务器开始监听...")
	listen, err := net.Listen("tcp", "localhost:8088")
	if err != nil {
		fmt.Println("listen err=,", err)
		return
	}
	defer listen.Close() //延时关闭

	//循环等待客户端来连接我
	for {
		fmt.Println("等待客户端连接")
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
	}

	//fmt.Printf("listen suc=%v", listen)
}

func process(conn net.Conn) {
	defer conn.Close()
	defer func() {
		mu.Lock()
		delete(connections, conn) // 移除连接
		mu.Unlock()
		conn.Close()
	}()
	//循环接收客户端发送的数据
	for {
		//创建一个新的切片
		buf := make([]byte, 1024)
		//等待客户端通过conn发送信息
		//如果客户端没有write，那么协程就会阻塞在这里
		//fmt.Printf("服务器在等待客户端%s发送信息\n", conn.RemoteAddr().String())
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("err=", err)
			return
		}
		//显示客户端发送的消息到终端
		//fmt.Print(string(buf[:n]))
		message := string(buf[:n])
		fmt.Print(message)

		// 广播消息
		mu.Lock()
		for c := range connections {
			if c != conn { // 不发送给自己
				_, _ = c.Write([]byte(message)) // 发送给其他客户端
			}
		}
		mu.Unlock()
	}
}
