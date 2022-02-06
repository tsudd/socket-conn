package utils

import "net"

type User struct {
	Name string
}

type Server struct {
	ConnectedUsers map[*User]net.UDPAddr
	E2EConnections map[*User][]*User
	WhiteList      map[string]*User
	Listener       *net.UDPConn
}
