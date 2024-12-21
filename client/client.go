package main

import (
	"bufio"
	"easy-chat/proto"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	userName string
	mu       sync.Mutex // 用于保护输入和消息显示的同步
)

const heartbeatInterval = 30 * time.Second // 心跳包发送间隔

func main() {
	//连接服务端
	conn, err := net.Dial("tcp", "localhost:8088")
	if err != nil {
		fmt.Println("Client connection to server failed, err=", err)
		return
	}
	defer conn.Close()

	//起始界面
	homeText() //起始界面

	reader := bufio.NewReader(conn)
	for {
		//填写昵称
		userName, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		userName = strings.Trim(userName, " \r\n")
		// 验证昵称
		if userName == "" {
			fmt.Println("昵称不能为空，请重新输入！")
			continue
		}
		//发送昵称到服务端
		data, err := proto.Encode(userName)
		if err != nil {
			fmt.Println("encode msg failed, err:", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("conn.Write err=", err)
		}

		msg, err := proto.Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("decode msg failed, err:", err)
			return
		}
		if msg == "true" {
			break
		} else {
			fmt.Println("昵称重复，请重新输入！")
			fmt.Println(" *请重新输入昵称↓↓↓")
			fmt.Printf(" >")
		}
	}

	//过渡动画
	loadText()
	//主界面
	mainText()

	// 开启心跳包发送协程
	go sendHeartbeat(conn)

	//接收服务端广播
	go func() {
		for {
			msg, err := proto.Decode(reader)
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Println("decode msg failed, err:", err)
				return
			}
			mu.Lock()
			// 使用 ANSI 转义序列移动光标
			fmt.Print("\033[G\033[K") // 移动光标到上一行并清除当前行
			fmt.Println(msg)          // 打印新消息
			fmt.Printf("> %s", "")    // 重新打印输入提示符
			mu.Unlock()
		}
	}()
	rd := bufio.NewReader(os.Stdin)
	//发送单行数据
	for {
		fmt.Print("> ")
		line, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println("readString err=", err)
			continue
		}
		line = strings.Trim(line, " \r\n")
		if line == "exit" || line == "EXIT" {
			fmt.Println("退出聊天室...")
			break
		}
		if line == "" {
			continue
		}
		//消息拼接
		date := time.Now().Format("[15:04:05]")
		massage := userName + date + ": " + line

		// 本地显示消息
		//mu.Lock()
		fmt.Printf("\033[1A\033[K") // 移动光标到上一行并清除当前行
		//fmt.Printf("%s\n", massage)
		//mu.Unlock()

		// 发送给服务器
		data, err := proto.Encode(massage)
		if err != nil {
			fmt.Println("encode msg failed, err:", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("conn.Write err=", err)
		}
		//fmt.Printf("发送%v字节数据", n)
	}
}

// sendHeartbeat 定期发送心跳包
func sendHeartbeat(conn net.Conn) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		msg := "###PING"
		data, err := proto.Encode(msg)
		if err != nil {
			fmt.Println("encode msg failed, err:", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("发送心跳包失败，可能已断开连接")
			return
		}
	}
}

// clearConsole 清空控制台
func clearConsole() {
	fmt.Print("\033[2J\033[3J") // 清除屏幕
	fmt.Print("\033[H")         // 将光标移动到左上角
}

// homeText 起始文本
func homeText() {
	clearConsole() //清空控制台
	fmt.Printf(`╔═══╗─────────────╔═══╗╔╗───────╔╗─────╔═══╗
║╔══╝─────────────║╔═╗║║║──────╔╝╚╗────║╔═╗║
║╚══╗╔══╗╔══╗╔╗─╔╗║║─╚╝║╚═╗╔══╗╚╗╔╝────║║─╚╝╔══╗
║╔══╝║╔╗║║══╣║║─║║║║─╔╗║╔╗║║╔╗║─║║─╔══╗║║╔═╗║╔╗║
║╚══╗║╔╗║╠══║║╚═╝║║╚═╝║║║║║║╔╗║─║╚╗╚══╝║╚╩═║║╚╝║
╚═══╝╚╝╚╝╚══╝╚═╗╔╝╚═══╝╚╝╚╝╚╝╚╝─╚═╝────╚═══╝╚══╝
─────────────╔═╝║─────by:RationalDysaniaer
─────────────╚══╝ 
`)
	fmt.Println("\n *欢迎来到EasyChat聊天室(^_^)/")
	fmt.Println(" *请输入昵称↓↓↓")
	fmt.Printf(" >")
}

// loadText 过渡动画
func loadText() {
	totalSteps := 40 // 进度条总长度
	var bar string
	fmt.Println()
	fmt.Printf("*正在进入聊天室，请稍等\n")
	for i := 0; i <= totalSteps; i++ {
		// 计算进度
		progress := float64(i) / float64(totalSteps)
		// 生成进度条字符串
		bar = "["
		for j := 0; j < i; j++ {
			bar = bar + "#"
		}
		for j := 0; j < totalSteps-i; j++ {
			bar = bar + "-"
		}
		bar = bar + "]"
		fmt.Printf("\r %s %.2f%%", bar, progress*100)
		time.Sleep(time.Millisecond * 60)
	}
	time.Sleep(time.Second / 4)
}

// mainText 聊天页面上方文本
func mainText() {
	clearConsole()
	fmt.Printf("EasyChat-Go    [currentUser:%v]\n", userName)
	fmt.Printf("-----------------------------------------\n")
}
