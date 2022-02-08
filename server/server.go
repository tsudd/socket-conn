package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tsudd/socket-conn/server/utils"
)

const (
	DefaultConfig              = "default_config.yaml"
	AlreadyConnectedUserAnswer = "User already connected to the server"
	ConnectedUsersListAnswer   = "Connected users:\n"
)

func main() {
	config := DefaultConfig
	if len(os.Args[1:]) > 0 {
		config = os.Args[1]
	}
	startServer(fmt.Sprintf("./config/%s", config))
}

func startServer(config string) {
	utils.LogMsg(fmt.Sprintf("Starting server using config from %s...", config))
	settings, err := utils.HandleConfig(config)
	utils.ChkErr(err)

	usersWithTokens := settings.Users

	srv := &utils.Server{
		ConnectedUsers: make(map[*utils.User]net.UDPAddr),
		E2EConnections: make(map[*utils.User][]*utils.User),
		WhiteList:      usersWithTokens,
	}

	utils.LogMsg("Init of UDP server with ", len(srv.WhiteList), " allowed users")

	listenAndHandle(srv, settings.Addr)
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
		go servConnection(srv, conn, buffer)
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

	switch message.Action {
	case utils.Ask:

	case utils.Con:
		if _, ok := srv.ConnectedUsers[user]; ok {
			utils.LogErr(token, " tries to connect again. aborting him.")
			go sendMessage(srv, con, buildServerMessage(AlreadyConnectedUserAnswer))
			return
		}
		userResponse := ConnectedUsersListAnswer
		var i = 1
		for username, addr := range srv.ConnectedUsers {
			userResponse += fmt.Sprintf("%d. %s on %s:%d\n", i, username, addr.IP.String(), addr.Port)
			i++
		}
		srv.ConnectedUsers[user] = *con
		go sendMessage(srv, con, buildServerMessage(userResponse))
	case utils.Mes:
		if chats_with, ok := srv.E2EConnections[user]; ok {
			for _, chatUser := range chats_with {
				if conn, ok := srv.ConnectedUsers[chatUser]; ok {
					utils.LogMsg("User ", token, "sends message to")
					go sendMessage(srv, &conn, message)
				}
			}
		}
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

func buildServerMessage(text string) utils.Message {
	return utils.Message{
		Action: utils.Srv,
		Params: map[string]string{
			utils.TimestampField: time.Now().String(),
			"text":               text,
		},
	}
}
