package src

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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

	user := NewUser(conn, this)

	//用户上线，将用户加入onlinemap
	user.Online()

	//监听用户是否活跃
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf) //读取用户发过来的消息
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			//用户针对msg进行消息处理
			user.DoMessage(msg)
			isLive <- true //用户的任意操作都代表目前活跃，重置定时器

		}
	}()

	//当前handler阻塞，不能关闭当前的handler
	for {
		select {
		//timer
		case <-isLive:
			//活跃，为了激活select，重置定时器
		case <-time.After(time.Second * 300):
			//关闭User
			user.sendMsg("CLOSED cause not active")
			user.Offline()
			//三件套：销毁-关闭-退出
			close(user.C)    //销毁资源
			conn.Close()     //关闭连接
			runtime.Goexit() //退出 用return也行
		}
	}

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
