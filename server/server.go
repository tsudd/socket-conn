package main

import (
	"errors"
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
	UnallowedUserAnswer        = "You are not allowed to use this server!"
	UnconnectedUserAnswer      = "You should connect before exchanging messages."
	UnallowedReceiverAnswer    = "This user doesn't exist in the system."
	UnconnectedRecieverAnswer  = "This user is disconnected. Try again later."
	ConnectionSuccesAnswer     = "Establishing process was succesful. You can start chatting soon."
	UserStartingChatAnswer     = "%s is starting chat with you. Say hi."
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
	settings, err := utils.HandleServerConfig(config)
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
		sendMessage(srv, con, buildServerMessage(UnallowedUserAnswer, utils.Err))
		return
	}

	switch message.Action {
	case utils.Ask:
		// check if user connected
		if _, ok := srv.ConnectedUsers[user]; !ok {
			utils.LogErr(token, " tries to exchange messages before connection. aboring him.")
			sendMessage(srv, con, buildServerMessage(UnconnectedUserAnswer, utils.Err))
			return
		}
		//check if other user exists in the system
		receiver, rtoken := checkUsername(srv, message.Params[utils.ReceiverTokenField])
		if receiver == nil {
			utils.LogErr(token, " tries to exchange messages with unallowed user. aborting him")
			sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
			return
		}
		// if other user connected trying to esablish connection
		if addr, ok := srv.ConnectedUsers[receiver]; ok {
			err := establishE2EConnection(srv, user, receiver)
			if err != nil {
				utils.LogErr("Abort establishing", err.Error())
				return
			}
			err = establishE2EConnection(srv, receiver, user)
			if err != nil {
				utils.LogErr("Abort establishing", err.Error())
				return
			}
			utils.LogMsg("Establishing connection between users", user.Name, " and ", receiver.Name)
			message.Params["text"] = fmt.Sprintf(UserStartingChatAnswer, user.Name)
			message.Params[utils.SenderNameField] = user.Name
			userResponse := buildServerMessage(ConnectionSuccesAnswer, utils.Ask)
			userResponse.Params[utils.TokenField] = rtoken
			message.Params[utils.SenderNameField] = receiver.Name
			go sendMessage(srv, &addr, message)
			go sendMessage(srv, con, userResponse)
		} else {
			utils.LogErr(token, " tries to exchange messages with unconnected user. aborting him")
			sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
		}
	case utils.Con:
		if _, ok := srv.ConnectedUsers[user]; ok {
			utils.LogErr(token, " tries to connect again. aborting him.")
			go sendMessage(srv, con, buildServerMessage(AlreadyConnectedUserAnswer, utils.Err))
			return
		}
		utils.LogMsg("Establishing connection with", user.Name)
		userResponse := ConnectedUsersListAnswer
		var i = 1
		for username, addr := range srv.ConnectedUsers {
			userResponse += fmt.Sprintf("%d. %s on %s:%d\n", i, username, addr.IP.String(), addr.Port)
			i++
		}
		srv.ConnectedUsers[user] = *con
		mes := buildServerMessage(userResponse, utils.Con)
		mes.Params["user"] = user.Name
		go sendMessage(srv, con, mes)
	case utils.Mes:
		if chatsWith, ok := srv.E2EConnections[user]; ok {
			receiver, ok := srv.WhiteList[message.Params[utils.ReceiverTokenField]]
			if !ok {
				utils.LogErr("User with ", message.Params[utils.ReceiverTokenField], " token is not allowed for messaging for ", user.Name)
				sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
				return
			}
			for _, chatUser := range chatsWith {
				addr, ok := srv.ConnectedUsers[receiver]
				if ok && receiver == chatUser {
					go sendMessage(srv, &addr, message)
					return
				}
			}
			utils.LogErr("User with ", message.Params[utils.ReceiverTokenField], " token is disconnected or unestablished for ", user.Name)
			go sendMessage(srv, con, buildServerMessage(UnconnectedRecieverAnswer, utils.Err))
			return
		}
	default:
		utils.LogMsg("undefined action from ", con.IP.String(), con.Port)
	}
}

func checkUsername(srv *utils.Server, name string) (*utils.User, string) {
	for t, user := range srv.WhiteList {
		if user.Name == name {
			return user, t
		}
	}
	return nil, ""
}

func sendMessage(srv *utils.Server, con *net.UDPAddr, message utils.Message) {
	buffer := utils.Enpack(message)
	_, err := srv.Listener.WriteToUDP(buffer, con)
	if err != nil {
		utils.LogErr("Couldn't send message to ", con.IP.String(), con.Port, err)
	}
}

func buildServerMessage(text string, action utils.Actions) utils.Message {
	return utils.Message{
		Action: action,
		Params: map[string]string{
			utils.TimestampField: time.Now().Format("2006-01-02 3:4:5 pm"),
			"text":               text,
		},
	}
}

func establishE2EConnection(srv *utils.Server, user1 *utils.User, user2 *utils.User) error {
	if connections, ok := srv.E2EConnections[user1]; ok {
		for _, con := range connections {
			if con == user2 {
				return errors.New("users already connected")
			}
		}
	}
	if _, ok := srv.E2EConnections[user1]; !ok {
		srv.E2EConnections[user1] = []*utils.User{user2}
	} else {
		srv.E2EConnections[user1] = append(srv.E2EConnections[user1], user2)
	}
	return nil
}

// func disconnectUser(srv *utils.Server, user *utils.User) {
// 	for u, connections := range srv.E2EConnections {
// 		if user == u {
// 			delete(srv.E2EConnections, u)
// 			continue
// 		}
// 		index := -1
// 		for i, con := range connections {
// 			if con == user {
// 				index = i
// 				break
// 			}
// 		}
// 		if index != -1 {
// 			srv.E2EConnections[u] = append(connections[:index], connections[index+1:]...)
// 		}
// 	}
// }
