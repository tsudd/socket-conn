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
	UnallowedUserAnswer        = "User with the same name already connected..."
	UnconnectedUserAnswer      = "You should connect before exchanging messages."
	UnallowedReceiverAnswer    = "This user doesn't exist in the system."
	UnconnectedRecieverAnswer  = "This user is disconnected. Try again later."
	ConnectionSuccesAnswer     = "Establishing process was succesful. You can start chatting soon."
	UserStartingChatAnswer     = "%s is starting chat with you. Say hi."
	UserReconnectedAnswer      = "%s is reconnected to the chat!\n"
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

	srv := &utils.Server{
		ConnectedUsers: make(map[*utils.User]*utils.Connection),
		E2EConnections: make(map[utils.User][]*utils.User),
		WhiteList:      make(map[string]*utils.User),
		Heartbeating:   15,
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
	tokenName := message.Params[utils.TokenField]
	user, ok := srv.WhiteList[tokenName]
	if !ok && message.Action == utils.Con {
		userToken := utils.GetMD5Hash(tokenName)
		user, ok = srv.WhiteList[userToken]
		if ok {
			connected, ok := srv.ConnectedUsers[user]
			if ok && -1*int(time.Until(connected.Alive).Seconds()) < srv.Heartbeating {
				utils.LogErr("User with name", tokenName, " already connected. Aborting ", con.IP.String(), con.Port)
				go sendMessage(srv, con, buildServerMessage(UnallowedUserAnswer, utils.Err))
				return
			}
		}
		userResponse := formConnectedUsersList(srv)
		user = &utils.User{Name: tokenName}
		srv.WhiteList[userToken] = user
		srv.ConnectedUsers[user] = &utils.Connection{Addr: *con, Alive: time.Now()}
		mes := buildServerMessage(userResponse, utils.Con)
		mes.Params[utils.SenderField] = userToken
		go sendMessage(srv, con, mes)
		utils.LogMsg("Establishing connection with", user.Name)
		return
	}
	if _, ok := srv.ConnectedUsers[user]; !ok {
		utils.LogErr(tokenName, " tries to exchange messages before connection. aboring him.")
		sendMessage(srv, con, buildServerMessage(UnconnectedUserAnswer, utils.Err))
		return
	}
	switch message.Action {
	case utils.Ask:
		//check if other user exists in the system
		receiver, rtoken := checkUsername(srv, message.Params[utils.ReceiverTokenField])
		if receiver == nil {
			utils.LogErr(tokenName, " tries to exchange messages with unallowed user. aborting him")
			sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
			return
		}
		if rtoken == tokenName {
			utils.LogErr(tokenName, " tries to exchange messages with himself. aborting him")
			sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
			return
		}
		// if other user connected trying to esablish connection
		if connection, ok := srv.ConnectedUsers[receiver]; ok {
			if ok && -1*int(time.Until(connection.Alive).Seconds()) > srv.Heartbeating {
				utils.LogErr("User ", receiver.Name, " was disconnected. Can't establish chatting. Aborting ", con.IP.String(), con.Port)
				go sendMessage(srv, con, buildServerMessage(UnconnectedRecieverAnswer, utils.Err))
				return
			}
			utils.LogMsg("Establishing connection between users", user.Name, " and ", receiver.Name)
			err := establishE2EConnection(srv, user, &receiver)
			if err != nil {
				utils.LogMsg("Abort establishing:", err.Error(), "Launching Sync process.")
				syncMessage := buildServerMessage(fmt.Sprintf(UserReconnectedAnswer, user.Name), utils.Syn)
				syncMessage.Params[utils.TokenField] = message.Params[utils.TokenField]
				syncMessage.Params[utils.SenderNameField] = user.Name
				go sendMessage(srv, &connection.Addr, syncMessage)
				return
			}
			err = establishE2EConnection(srv, receiver, &user)
			if err != nil {
				utils.LogErr("Abort establishing:", err.Error())
				return
			}
			message.Params["text"] = fmt.Sprintf(UserStartingChatAnswer, user.Name)
			message.Params[utils.SenderNameField] = user.Name
			userResponse := buildServerMessage(ConnectionSuccesAnswer, utils.Ask)
			userResponse.Params[utils.TokenField] = rtoken
			message.Params[utils.SenderNameField] = receiver.Name
			go sendMessage(srv, &connection.Addr, message)
			go sendMessage(srv, con, userResponse)
		} else {
			utils.LogErr(tokenName, " tries to exchange messages with unconnected user. aborting him")
			sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
		}
	case utils.Con:
		userResponse := formConnectedUsersList(srv)
		mes := buildServerMessage(userResponse, utils.Srv)
		srv.ConnectedUsers[user].Alive = time.Now()
		go sendMessage(srv, con, mes)
	case utils.Mes:
		if chatsWith, ok := srv.E2EConnections[*user]; ok {
			receiver, ok := srv.WhiteList[message.Params[utils.ReceiverTokenField]]
			if !ok {
				utils.LogErr("User with ", message.Params[utils.ReceiverTokenField], " token is not allowed for messaging for ", user.Name)
				sendMessage(srv, con, buildServerMessage(UnallowedReceiverAnswer, utils.Err))
				return
			}
			connection, ok := srv.ConnectedUsers[receiver]
			for _, chatUser := range chatsWith {
				if ok && receiver == chatUser {
					go sendMessage(srv, &connection.Addr, message)
					return
				}
			}
			utils.LogErr("User with ", message.Params[utils.ReceiverTokenField], " token is disconnected or unestablished for ", user.Name)
			srv.ConnectedUsers[user].Alive = time.Now()
			go sendMessage(srv, con, buildServerMessage(UnconnectedRecieverAnswer, utils.Err))
			return
		}
	case utils.Syn:
		receiver := srv.WhiteList[message.Params[utils.ReceiverTokenField]]
		if connection, ok := srv.ConnectedUsers[receiver]; ok {
			message.Action = utils.Ask
			go sendMessage(srv, &connection.Addr, message)
		}
	case utils.Hrt:
		srv.ConnectedUsers[user].Alive = time.Now()
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

func establishE2EConnection(srv *utils.Server, user1 *utils.User, user2 **utils.User) error {
	connections, ok := srv.E2EConnections[*user1]
	if ok {
		for _, con := range connections {
			if con == *user2 {
				return errors.New("users already connected")
			}
		}
		srv.E2EConnections[*user1] = append(srv.E2EConnections[*user1], *user2)
	} else {
		srv.E2EConnections[*user1] = []*utils.User{*user2}
	}
	return nil
}

func formConnectedUsersList(srv *utils.Server) string {
	if len(srv.ConnectedUsers) == 0 {
		return "Server is empty...\n"
	}
	userResponse := ConnectedUsersListAnswer
	i := 1
	for username, conn := range srv.ConnectedUsers {
		userResponse += fmt.Sprintf("%d. %s on %s:%d\n", i, username.Name, conn.Addr.IP.String(), conn.Addr.Port)
		i++
	}
	return userResponse
}
