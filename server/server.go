package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/tsudd/socket-conn/server/utils"
)

const (
	DefaultConfig = "default_config.yaml"
)

type Config struct {
	addr net.UDPAddr
}

func main() {
	config := DefaultConfig
	if len(os.Args[1:]) > 0 {
		config = os.Args[1]
	}
	startServer(fmt.Sprintf("./config/%s", config))
}

func startServer(config string) {
	utils.LogMsg(fmt.Sprintf("Starting server using config from %s...", config))
	settings, err := handleConfig(config)
	utils.ChkErr(err)
	listen, err := net.ListenUDP("udp4", &settings.addr)
	utils.ChkErr(err)
	defer listen.Close()
	buffer := make([]byte, 2048)
	utils.LogMsg("Waiting for clients")

	for {
		_, conn, err := listen.ReadFromUDP(buffer)
		utils.LogMsg(conn.IP.String(), conn.Port, " read message ", string(buffer))
		if err != nil {
			continue
		}

		// go handleConnection(conn, 5)
	}
}

func handleConfig(path string) (Config, error) {
	configs := utils.GetConfig(path)
	port, err := strconv.Atoi(utils.GetElement("port", configs))
	if err != nil {
		return Config{}, err
	}
	return Config{
		addr: net.UDPAddr{
			IP:   net.ParseIP(utils.GetElement("IP", configs)),
			Port: port,
			Zone: "",
		},
	}, nil

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
		tempBuf = utils.Depack(append(tempBuf, buffer[:n]...))
		utils.LogMsg("Received data: ", string(tempBuf))

		//start heartbeating
		go utils.HeartBeating(conn, messanger, timeout)
		//check if get message from client
		go utils.GravelChannel(tempBuf, messanger)
	}
	defer conn.Close()
}
