package utils

import (
	"net"
	"time"
)

type User struct {
	Name string
}

type ClientMode int

const (
	Chat ClientMode = iota
	Wait
)

type Connection struct {
	Addr  net.UDPAddr
	Alive time.Time
}

type Server struct {
	ConnectedUsers map[*User]*Connection
	E2EConnections map[User][]*User
	WhiteList      map[string]*User
	Listener       *net.UDPConn
	Heartbeating   int
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
