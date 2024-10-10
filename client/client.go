package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8088")
	if err != nil {
		fmt.Println("client dial err=", err)
		return
	}
	defer conn.Close()

	go func() {
		for {
			message := make([]byte, 1024)
			n, err := conn.Read(message)
			if err != nil {
				fmt.Println("服务器连接已关闭")
				return
			}
			fmt.Print(string(message[:n]))
		}
	}()

	//发送单行数据
	reader := bufio.NewReader(os.Stdin) //os.Stdin 代表标准输入（终端）

	for {
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
		_, err = conn.Write([]byte(line + "\n"))
		if err != nil {
			fmt.Println("conn.Write err=", err)
		}
		//fmt.Printf("发送%v字节数据", n)
	}

}
