package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	userName string
	mu       sync.Mutex // 用于保护输入和消息显示的同步
)

func main() {
	//连接服务端
	conn, err := net.Dial("tcp", "localhost:8088")
	if err != nil {
		fmt.Println("client dial err=", err)
		return
	}
	defer conn.Close()

	//填写昵称
	reader := bufio.NewReader(os.Stdin) //os.Stdin 代表标准输入（终端）
	fmt.Print("\033[2J\033[3J")         // 清除屏幕
	fmt.Print("\033[H")                 // 将光标移动到左上角
	fmt.Println("欢迎来到EasyChat聊天室！")
	fmt.Printf("请输入昵称:")
	userName, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("readString err=", err)
	}
	userName = strings.Trim(userName, " \r\n")
	fmt.Println("已进入聊天室，请尽情畅聊！！！")
	fmt.Println("------------------------------------------")

	//接收服务端广播
	go func() {
		for {
			message := make([]byte, 1024)
			n, err := conn.Read(message)
			if err != nil {
				fmt.Println("服务器连接已关闭")
				return
			}
			//fmt.Print(string(message[:n]))
			msg := string(message[:n])

			// 锁定用于同步
			mu.Lock()
			// 使用 ANSI 转义序列移动光标
			fmt.Printf("\033[G\033[K") // 移动光标到上一行并清除当前行
			fmt.Print(msg)             // 打印新消息
			fmt.Printf("> %s", "")     // 重新打印输入提示符
			mu.Unlock()
		}
	}()

	//发送单行数据
	for {
		//fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("readString err=", err)
			continue
		}
		line = strings.Trim(line, " \r\n")
		if line == "exit" {
			fmt.Println("客户端退出..")
			break
		}

		// 锁定，用于同步显示
		mu.Lock()
		// 发送并本地显示消息
		fmt.Printf("\033[1A\033[K") // 移动光标到上一行并清除当前行
		fmt.Printf("%s: %s\n", userName, line)
		fmt.Printf("> %s", "") // 再次打印输入提示符
		mu.Unlock()

		// 发送给服务器
		_, err = conn.Write([]byte(userName + ":" + line + "\n"))
		if err != nil {
			fmt.Println("conn.Write err=", err)
		}
		//fmt.Printf("发送%v字节数据", n)
	}
}
