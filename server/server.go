package main

import (
	"fmt"
	"net"
	"os"
	"server/utils"
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
	}
}
