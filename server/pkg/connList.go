package pkg

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// ConnList 连接列表
type ConnList struct {
	Connections map[net.Conn]*connState
	rw          sync.RWMutex // 保护连接的读写
}

// connState 连接状态
type connState struct {
	NickName      string
	Add           string
	LoginTime     time.Time
	LastHeartTime time.Time
}

// CreatConnList 连接列表初始化
func CreatConnList() *ConnList {
	return &ConnList{
		Connections: make(map[net.Conn]*connState),
		rw:          sync.RWMutex{},
	}
}

// Add 添加客户端连接
func (c *ConnList) Add(conn net.Conn, nickName string) {
	state := &connState{
		NickName:      nickName,
		Add:           conn.RemoteAddr().String(),
		LoginTime:     time.Now(),
		LastHeartTime: time.Now(),
	}
	c.rw.Lock()
	c.Connections[conn] = state
	c.rw.Unlock()
}

// Delete 删除客户端连接
func (c *ConnList) Delete(conn net.Conn) {
	c.rw.Lock()
	delete(c.Connections, conn)
	c.rw.Unlock()
}

// GetList 用户连接列表
func (c *ConnList) GetList() string {
	var message string
	message = message + "---------------------------------------------------\n当前用户列表：\n"
	message = message + fmt.Sprintf("IP              登录时间            昵称\n")
	for n, v := range c.Connections {
		message = message + fmt.Sprintf("%v %v %v\n", n.RemoteAddr().String(), v.LoginTime.Format("2006:01:02 15:04:05"), v.NickName)
	}
	message = message + "---------------------------------------------------"
	return message
}

// GetLastHeardTime 显示心跳时间
func (c *ConnList) GetLastHeardTime() string {
	var message string
	message = message + "---------------------------------------------------\n用户心跳列表：\n"
	message = message + fmt.Sprintf("登录时间            最后心跳时间        昵称\n")
	for _, v := range c.Connections {
		message = message + fmt.Sprintf("%v %v %v\n", v.LoginTime.Format("2006:01:02 15:04:05"), v.LastHeartTime.Format("2006:01:02 15:04:05"), v.NickName)
	}
	message = message + "---------------------------------------------------"
	return message
}

// IsExist 连接是否存在
func (c *ConnList) IsExist(conn net.Conn) bool {
	c.rw.RLock()
	_, exists := c.Connections[conn]
	c.rw.RUnlock()
	return exists
}

// IsNameExist 昵称是否已存在
func (c *ConnList) IsNameExist(nickName string) bool {
	for _, v := range c.Connections {
		if v.NickName == nickName {
			return true
		}
	}
	return false
}

// GetConnByNickName 通过昵称获取连接
func (c *ConnList) GetConnByNickName(nickName string) (net.Conn, error) {
	for k, v := range c.Connections {
		if v.NickName == nickName {
			return k, nil
		}
	}
	return nil, errors.New("no user")
}

// GetAllConn 获取所有连接
func (c *ConnList) GetAllConn() map[net.Conn]*connState {
	return c.Connections
}
