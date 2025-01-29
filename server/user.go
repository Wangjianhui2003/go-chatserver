package server

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	conn   net.Conn
	C      chan string
	server *Server
}

// 新建用户
func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		conn:   conn,
		C:      make(chan string),
		server: server,
	}

	//go程监听user的C
	go user.ListenUserMessage()

	return user
}

// 监听用户的channel,写入连接
func (this *User) ListenUserMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线")
}

func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "下线")
}

func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {
	if msg == "who" { //查询在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			this.sendMsg("[" + user.Addr + "]" + user.Name + "在线\n")
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { //重命名
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]

		if ok {
			this.sendMsg("用户名已存在\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			this.Name = newName
			this.sendMsg("用户名已修改\n")
		}
	} else if len(msg) > 3 && msg[:3] == "to|" { //私聊
		remoteUserName := strings.Split(msg, "|")[1]
		if remoteUserName == "" {
			this.sendMsg("格式不正确,请输入to|name|content\n")
			return
		}

		remoteUser, ok := this.server.OnlineMap[remoteUserName]
		if !ok {
			this.sendMsg("用户名不存在")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("消息为空!")
			return
		}

		remoteUser.sendMsg("[私聊]" + this.Name + ":" + content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}
