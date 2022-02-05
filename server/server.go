package main

import (
	"fmt"
	"net"
	"os"

	"github.com/tsudd/socket-conn/server/utils"
)

const (
	defaultConfig = "default_config.yaml"
)

func main() {
	config := defaultConfig
	if len(os.Args[1:]) > 0 {
		config = os.Args[1]
	}
	startServer(fmt.Sprintf("./config/%s", config))
}

func startServer(config string) {
	utils.LogMsg(fmt.Sprintf("Starting server using config from %s...", config))
	configs := utils.GetConfig(config)
	host := utils.GetElement("host", configs)
	listen, err := net.Listen("tcp", host)
	utils.ChkErr(err)
	defer listen.Close()
	utils.LogMsg("Waiting for clients")

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		utils.LogMsg(conn.RemoteAddr().String(), " tcp connect success")
		go handleConnection(conn, 5)
	}
}

func handleConnection(conn net.Conn, timeout int) {
	tempBuf := make([]byte, 0)
	buffer := make([]byte, 1024)
	messanger := make(chan byte)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			utils.LogErr(conn.RemoteAddr().String(), " connection error", err)
			return
		}
		tempBuf = append(tempBuf, buffer[:n]...)
		utils.LogMsg("Received data: ", string(tempBuf))

		//start heartbeating
		go utils.HeartBeating(conn, messanger, timeout)
		//check if get message from client
		go utils.GravelChannel(tempBuf, messanger)
	}
	defer conn.Close()
}
