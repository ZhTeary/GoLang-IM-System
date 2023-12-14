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
	flag       int //当前客户端的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       9,
	}

	//连接服务器！
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}
	client.conn = conn
	//返回对象
	return client
}

// 处理服务器返回给客户端的消息
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)

}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊")
	fmt.Println("2.私聊")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>>>>>请输入合法范围内的数字<<<<<<<<<<")
		return false
	}

}

func (client *Client) UpdateName() bool {
	fmt.Println(">>> 请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"

	//将客户端消息发给服务器
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (client *Client) PublicChat() {

	fmt.Println(">>>>>请输入聊天内容,exit退出")
	//输入消息
	var chatMsg string
	fmt.Scanln(&chatMsg)

	//发给服务器
	for chatMsg != "exit" {

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)

	}

	//从服务器返回的消息有DealResponse接收
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// 私聊模式
func (client *Client) SideText() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println(">>>>>>>>>>>请输入聊天对象[用户名],exit退出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>> 请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"

				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>> 请输入消息内容，exit退出")
			fmt.Scanln(&chatMsg)

		}
	}
	client.SelectUsers()
	fmt.Println(">>>>>>>>>>>请输入聊天对象[用户名],exit退出：")
	fmt.Scanln(&remoteName)

}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		//根据不同模式，处理不同业务
		switch client.flag {
		case 1: //公聊
			fmt.Println("公聊模式选择.......")
			client.PublicChat()
			break

		case 2: //私聊
			fmt.Println("私聊模式选择.......")
			client.SideText()
			break
		case 3: //更新用户名
			fmt.Println("更新用户名选择.......")
			client.UpdateName()
			break
		}

	}
}

var serverIP string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认8888)")

}

func main() {
	// 解析命令行
	flag.Parse()

	//向服务器发送
	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>连接服务器失败....")
		return
	}

	//单独开启一个goroutine处理server的回执消息
	go client.DealResponse()

	fmt.Println(">>>>>>>>连接服务器成功....")

	//启动客户端业务
	client.Run()
}
