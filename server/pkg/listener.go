package pkg

import (
	"errors"
	"net"
)

// MyListener 监听器
type MyListener struct {
	listener net.Listener
}

func CreateListener() *MyListener {
	listener := &MyListener{}
	return listener
}

// Close 关闭监听
func (m *MyListener) Close() {
	_ = m.listener.Close()
}

// StartListen 开始监听
func (m *MyListener) StartListen(address string) error {
	var err error
	m.listener, err = net.Listen("tcp", address)
	if err != nil {
		return errors.New("listen err=" + err.Error())
	}
	return nil
}

func (m *MyListener) Accept() (net.Conn, error) {
	return m.listener.Accept()
}
