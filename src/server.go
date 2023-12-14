package src

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// //+ userMap
	OnlineMap map[string]*User
	maplock   sync.RWMutex

	// //+ messagechan用于广播
	Message chan string
}

// 创建server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (this *Server) Handler(conn net.Conn) {
	//..当前连接的业务
	fmt.Println("OKOKOK")

	user := NewUser(conn)

	// //用户上线，将用户加入onlinemap
	this.maplock.Lock()
	this.OnlineMap[user.Name] = user
	this.maplock.Unlock()

	//广播消息
	this.BroadCast(user, "Online NOW!")

	//当前handler阻塞
	select {}
}

// 广播消息内容
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg

}

// 监听Message广播消息channel的goroutine，一旦有消息就广播给全部在线client
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//将所有msg发送给全部的在线user
		this.maplock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.maplock.Unlock()
	}

}

// 启动服务器
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net listener err :", err)
		return
	}
	defer listener.Close()

	//启动监听Messgae的goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//accept之后代表用户上线
		//do handler
		go this.Handler(conn)

	}

	//close listener socket
}
