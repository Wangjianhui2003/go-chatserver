package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

// 新建客户端
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&flag)
	if err != nil {
		return false
	}

	if flag >= 0 && flag <= 4 {
		client.flag = flag
	} else {
		fmt.Println("命令不合法")
		return false
	}
	return true
}

// 更新用户名
func (client *Client) UpdateName() bool {
	fmt.Println(">>>请输入用户名")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn Write error", err)
		return false
	}

	return true
}

func (client *Client) PublicTalk() {
	fmt.Println(">>>公聊模式,输入exit退出")
	var chatMsg string
	fmt.Scanln(&chatMsg)

	//不是exit
	for chatMsg != "exit" {
		//消息不为空
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err", err)
				break
			}
		}

		//循环
		chatMsg = ""
		//fmt.Println(">>>公聊模式,输入exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) PrivateTalk() {
	var remoteName string
	var chatMsg string

	fmt.Println(">>>私聊模式,输入用户名进行私聊")
	client.ListUsers()
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("---输入消息:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("---")
			fmt.Scanln(&chatMsg)
		}

		client.ListUsers()
		fmt.Println(">>>私聊模式,输入用户名进行私聊")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) ListUsers() {
	sendMsg := "who\n"

	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn Write err", err)
		return
	}
}

// 显示服务器发来的消息
func (client *Client) ShowServerMsg() {
	//一直阻塞
	io.Copy(os.Stdout, client.conn)

	//for {
	//	buf := make([]byte,4096)
	//	client.conn.Read(buf)
	//	fmt.Println(buf)
	//}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		switch client.flag {
		case 1:
			client.PublicTalk()
			break
		case 2:
			client.PrivateTalk()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// 读取命令行参数
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip")
	flag.IntVar(&serverPort, "port", 8001, "设置服务器port")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println("连接失败")
	}

	fmt.Println(">>>>>>>>连接成功<<<<<<<<<")
	//开启一个go程
	go client.ShowServerMsg()

	client.Run()
}
