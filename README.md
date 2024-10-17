# Easy-Chat

**项目名称**：Easy-Chat
**开发语言**：Go
**项目描述**：基于命令行的多人聊天室，用来练习go基础知识



## 功能特点：

- 多客户端支持：多个用户可同时连接到服务端，进行实时交流。
- 用户昵称设置：客户端启动后可以设置昵称，昵称会在聊天室中显示。
- 消息广播：客户端发送的消息会被广播给所有在线用户。
- 心跳检测机制：服务端定时检测客户端的连接状态，若客户端未及时发送心跳包，连接将被断开。
- 简洁用户界面：使用命令行界面，显示当前在线用户及实时消息。

项目结构：

```
easy-chat/
│
├── client/
│   └── client.go        # 客户端实现
│
├── server/
│   └── server.go        # 服务端实现
│
├── go.mod               # Go 依赖模块管理文件
├── LICENSE              # 许可证文件 (Apache License 2.0)
└── README.md            # 项目说明文档
```



## 安装与使用步骤

### 环境要求

- Go 1.16 及以上版本

### 运行步骤

1. 克隆项目到本地：

   ```
   git clone https://github.com/your-repo/easy-chat.git
   cd easy-chat
   ```

2. 启动服务端：

   ```
   cd server
   go run server.go
   ```

   服务端将在 localhost:8088 端口上监听客户端的连接。

3. 启动客户端

   在新的终端窗口中进入客户端目录：

   ```
   cd client
   go run client.go
   ```

   客户端运行后，将提示输入昵称，输入后即可加入聊天室进行聊天。



## 注意事项

- 请确保客户端和服务端在同一台机器或网络中，或确保防火墙开放了相应的 TCP 端口（默认为 `8088`）。
- 心跳机制是为了检测连接的健康状态，如果客户端由于网络问题长时间没有响应，服务端会自动断开连接。



## 贡献

欢迎提交 Pull Request 或 Issue 来改进该项目！如果有任何问题或建议，欢迎提问。



## 项目许可证：

该项目基于 Apache License 2.0 许可证开源。详情请查看 LICENSE 文件。



## 开发者：

本项目由 RationalDysaniaer 开发。

