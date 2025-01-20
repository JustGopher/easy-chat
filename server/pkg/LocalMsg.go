package pkg

import (
	"fmt"
	"sync"
)

// LocalMsg 本地消息
type LocalMsg struct {
	msg chan string
	mu  sync.Mutex
}

// CreateLocalMsg 创建本地消息实例
func CreateLocalMsg() *LocalMsg {
	return &LocalMsg{
		msg: make(chan string),
	}
}

// Add 添加消息
func (l *LocalMsg) Add(s string) {
	l.mu.Lock()
	l.msg <- s
	l.mu.Unlock()
}

// Out 输出消息
func (l *LocalMsg) Out() {
	for s := range l.msg {
		fmt.Print("\033[G\033[K")
		fmt.Println(s)
		fmt.Print("> ")
	}
}

// ClearConsole 清空控制台
func (l *LocalMsg) ClearConsole() {
	l.Add("\033[2J\033[3J") // 清除屏幕
	l.Add("\033[H")         // 将光标移动到左上角
}

// HomeText 起始界面
func (l *LocalMsg) HomeText() {
	l.ClearConsole()
	l.Add(`╔═══╗─────────────╔═══╗╔╗───────╔╗─────╔═══╗
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
