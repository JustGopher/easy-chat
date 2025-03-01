package pkg

import (
	"easy-chat/proto"
	"errors"
	"net"
)

// BroadcastMsg 广播消息
type BroadcastMsg struct {
	msg chan string
}

// CreateBroadcastMsg 创建广播消息处理
func CreateBroadcastMsg() *BroadcastMsg {
	return &BroadcastMsg{
		msg: make(chan string),
	}
}

// Add 添加广播消息
func (bc *BroadcastMsg) Add(message string) {
	bc.msg <- message
}

// SendMessage 发送广播消息
func (bc *BroadcastMsg) SendMessage(conn map[net.Conn]*connState) error {
	for message := range bc.msg {
		for c, _ := range conn {
			data, err := proto.Encode(message)
			if err != nil {
				return errors.New("encode msg failed, go: sendMessage(), err=" + err.Error())
			}
			_, err = c.Write(data)
			if err != nil {
				return errors.New("sendMessage failed, go: sendMessage(), err=" + err.Error())
			}
		}
	}
	return nil
}
