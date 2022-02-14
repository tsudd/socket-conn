package utils

import "net"

type User struct {
	Name string
}

type ClientMode int

const (
	Chat ClientMode = iota
	Wait
)

type Server struct {
	ConnectedUsers map[*User]net.UDPAddr
	E2EConnections map[*User][]*User
	WhiteList      map[string]*User
	Listener       *net.UDPConn
}

type Client struct {
	User          *User
	UserToken     string
	ServerAddr    net.UDPAddr
	Dial          *net.UDPConn
	ReceiverToken string
	ReceiverName  string
	Mode          ClientMode
	History       []string
}
