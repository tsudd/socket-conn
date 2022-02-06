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
	addr  net.UDPAddr
	users map[string]*utils.User
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

	usersWithTokens := settings.users

	srv := &utils.Server{
		ConnectedUsers: make(map[*utils.User]net.UDPAddr),
		E2EConnections: make(map[*utils.User][]*utils.User),
		WhiteList:      usersWithTokens,
	}

	listenAndHandle(srv, settings.addr)
}

func listenAndHandle(srv *utils.Server, addr net.UDPAddr) error {
	listen, err := net.ListenUDP("udp4", &addr)
	utils.ChkErr(err)
	defer listen.Close()
	srv.Listener = listen

	buffer := make([]byte, 2048)
	// messanger := make(chan byte)
	utils.LogMsg("Waiting for clients")

	for {
		_, conn, err := listen.ReadFromUDP(buffer)
		if err != nil {
			continue
		}
		utils.LogMsg(conn.IP.String(), conn.Port, " read message ", string(buffer))
		go servConnection(srv, conn, buffer)
		// go utils.HeartBeating(*listen, messanger, 5)
		// go utils.GravelChannel(buffer, messanger)
		// go handleConnection(conn, 5)
	}
}

func servConnection(srv *utils.Server, con *net.UDPAddr, buffer []byte) {
	message := utils.Depack(buffer)
	token := message.Params[utils.TokenField]
	user, ok := srv.WhiteList[token]
	if !ok {
		utils.LogErr("Wrong user with token", token, " from ", con.IP.String(), con.Port)
		return
	}

	srv.ConnectedUsers[user] = *con
	// message = utils.Depack(srv.Listener.ReadFromUDP())
	switch message.Action {
	case utils.Mes:
		utils.LogMsg(con.IP.String(), con.Port, " read message ", message.Params["text"])
		go sendMessage(srv, con, message)
	default:
		utils.LogMsg("undefined action from ", con.IP.String(), con.Port)
	}
}

func sendMessage(srv *utils.Server, con *net.UDPAddr, message utils.Message) {
	buffer := utils.Enpack(message)
	_, err := srv.Listener.WriteToUDP(buffer, con)
	if err != nil {
		utils.LogErr("Couldn't send message to ", con.IP.String(), con.Port, err)
	}
}

func handleConfig(path string) (Config, error) {
	configs := utils.GetConfig(path)
	port, err := strconv.Atoi(utils.GetElement("port", configs))
	if err != nil {
		return Config{}, err
	}
	// users := utils.GetElement("users", configs)
	// usersWithTokens := make(map[string]*utils.User)
	// for toke, user := range users {

	// }
	return Config{
		addr: net.UDPAddr{
			IP:   net.ParseIP(utils.GetElement("IP", configs)),
			Port: port,
			Zone: "",
		},
		users: map[string]*utils.User{
			"387e6278d8e06083d813358762e0ac63": {
				Name: "joohncena",
			},
		},
	}, nil

}

// func handleConnection(conn net.Conn, timeout int) {
// 	tempBuf := make([]byte, 0)
// 	buffer := make([]byte, 1024)
// 	messanger := make(chan byte)
// 	for {
// 		n, err := conn.Read(buffer)
// 		if err != nil {
// 			utils.LogErr(conn.RemoteAddr().String(), " connection error", err)
// 			return
// 		}
// 		tempBuf = utils.Depack(append(tempBuf, buffer[:n]...))
// 		utils.LogMsg("Received data: ", string(tempBuf))

// 		//start heartbeating
// 		go utils.HeartBeating(conn, messanger, timeout)
// 		//check if get message from client
// 		go utils.GravelChannel(tempBuf, messanger)
// 	}
// 	defer conn.Close()
// }
