package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// 服务器对象
type Server struct {
	Ip   string
	Port string

	//用户map mapLock为零值也可以使用
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播chan
	Message chan string
}

// 新建服务器
func NewServer(ip, port string) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 启动服务器接口
func (this *Server) Start() {
	fmt.Println("服务器启动")
	//监听对象
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	defer listener.Close()

	//监听Server的Message
	go this.ListenServerMessage()

	//获得连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
			continue
		}

		//开一个go程处理该连接
		go this.NewHandler(conn)
	}
}

// 连接后的处理逻辑
func (this *Server) NewHandler(conn net.Conn) {
	user := NewUser(conn, this)
	user.Online()

	//开一个go程接收消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read Error:", err)
				return
			}

			//广播消息
			msg := string(buf[:n-1]) //去掉\n
			user.DoMessage(msg)
		}
	}()

	select {}
}

// 发送广播消息到Message
func (this *Server) BroadCast(user *User, msg string) {
	sendmsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendmsg
}

// 监听Server的Message,发给每一个User
func (this *Server) ListenServerMessage() {
	for {
		sendmsg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- sendmsg
		}
		this.mapLock.Unlock()
	}
}
