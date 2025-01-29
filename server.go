package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
		go this.Handle(conn)
	}
}

// 连接后的处理逻辑
func (this *Server) Handle(conn net.Conn) {
	user := NewUser(conn, this)
	user.Online()
	isLive := make(chan bool)
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

			//处理用户发来的信息
			isLive <- true
			msg := string(buf[:n-1]) //去掉\n
			user.DoMessage(msg)
		}
	}()

	//超时强踢
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 120): //其他的case执行后定时器会刷新
			user.sendMsg("你被踢了")
			//关闭user channel
			close(user.C)
			//关闭连接
			conn.Close()
			return
		}
	}
}

// 发送广播消息到Message
func (this *Server) BroadCast(user *User, msg string) {
	sendmsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendmsg
}

// 监听Server的Message,发给每一个User的channel
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
