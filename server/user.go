package server

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 监听当前User channel 的Go方法，一旦有消息就发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}

}

// 用户上线
func (this *User) Online() {

	this.server.maplock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.maplock.Unlock()

	//广播消息
	this.server.BroadCast(this, "online")

}

// 用户下线
func (this *User) Offline() {

	this.server.maplock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.maplock.Unlock()

	this.server.BroadCast(this, "offline")

}

// v0.5 给当前用户对应的客户端发消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户
		this.server.maplock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..\n "
			this.sendMsg(onlineMsg)
		}
		this.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式"rename|张三"
		newName := strings.Split(msg, "|")[1] //标志在0，后面的在1

		//name是否存在
		if _, ok := this.server.OnlineMap[newName]; ok {
			this.sendMsg("当前用户名已经被使用")
		} else {
			this.server.maplock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.maplock.Unlock()

			this.Name = newName
			this.sendMsg("您已经更新用户名:" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式： to|张三|消息内容
		//step1. 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.sendMsg("消息格式不正确，请使用\"to|张三|hello\"格式\n")
		}
		//step2. 根据对方名，得到对方的user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.sendMsg("该用户名不存在\n")
			return
		}

		//step3. 获取消息内容，通过对方的user对象将内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("无消息内容，请重新发送\n")
			return
		}
		remoteUser.sendMsg(this.Name + "对您说：" + content)
	} else {

		this.server.BroadCast(this, msg)

	}
}
