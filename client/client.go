package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tsudd/socket-conn/server/utils"
)

const (
	StartChatCommand = "chat"
	HelpCommand      = "help"
)

const (
	HelpMessage = "Help: commands:\nchat <username> - start chatting with specific user, user should be connected to the server\nexit - end client programm\nhelp - get help\n"
)

func main() {
	var config string
	if len(os.Args[1:]) > 0 {
		config = os.Args[1]
	}
	if len(config) == 0 {
		utils.LogErr("Please provide config file to use client!")
		return
	}
	startClient(fmt.Sprintf("./config/%s", config))
}

func startClient(path string) {
	fmt.Printf("Starting client using config from %s...\n", path)
	settings, err := utils.HandleClientConfig(path)
	utils.ChkErr(err)

	cln := &utils.Client{
		UserToken:  settings.UserToken,
		ServerAddr: settings.Addr,
		Mode:       utils.Wait,
	}

	fmt.Println("Init of UDP client with ", cln.UserToken, " token for server ", cln.ServerAddr.IP.String(), ":", cln.ServerAddr.Port)
	fmt.Printf("Starting connection to %s:%d....\n", cln.ServerAddr.IP.String(), cln.ServerAddr.Port)
	err = getConnection(cln)
	if err != nil {
		utils.ChkErr(err)
		os.Exit(0)
	}
	fmt.Println("Succesfully connected to the server!")
	go handleReading(cln)
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Err occured while reading input...")
		}

		if cln.Mode == utils.Wait {
			inputs := strings.Split(input[0:len(input)-1], " ")
			argAmount := len(inputs)
			if argAmount == 0 || argAmount > 2 {
				fmt.Printf("Wrong command format. Use %s to see usages.", HelpCommand)
				continue
			}
			command := inputs[0]
			switch command {
			case StartChatCommand:
				if argAmount == 1 {
					fmt.Printf("Wrong input format for %s command. Use %s to see usage of exising commands.\n", StartChatCommand, HelpCommand)
					fmt.Println(input)
					break
				}
				mes := buildClientMessage(cln, utils.Ask)
				mes.Params[utils.ReceiverTokenField] = inputs[1]
				mes.Params[utils.SenderNameField] = cln.User.Name
				sendMessage(cln, mes)
			case HelpCommand:
				fmt.Print(HelpMessage)
			default:
				fmt.Printf("Unknown command...\n")
			}
		} else {
			mes := buildClientMessage(cln, utils.Mes)
			mes.Params[utils.ReceiverTokenField] = cln.ReceiverToken
			mes.Params["text"] = input
			go sendMessage(cln, mes)
		}
	}
}

func handleReading(cln *utils.Client) {
	fmt.Println("Starting reading from server...")
	for {
		mes, err := receiveMessage(cln, 15)
		if err != nil {
			fmt.Printf("Error while reading from the server: %s", err.Error())
			continue
		}
		switch mes.Action {
		case utils.Ask:
			cln.Mode = utils.Chat
			cln.ReceiverToken = mes.Params[utils.TokenField]
			cln.ReceiverName = mes.Params[utils.SenderNameField]
		default:

		}
		outputMessage(cln, mes)
	}
}

func getConnection(cln *utils.Client) error {
	conn, err := net.DialUDP("udp4", nil, &cln.ServerAddr)
	if err != nil {
		return err
	}
	cln.Dial = conn
	connRequest := buildClientMessage(cln, utils.Con)
	sendMessage(cln, connRequest)
	mes, err := receiveMessage(cln, 5)
	if err != nil {
		return err
	}
	if mes.Action == utils.Con {
		cln.User = &utils.User{Name: mes.Params["user"]}
		fmt.Print(mes.Params["text"])
		return nil
	} else {
		return errors.New(mes.Params["text"])
	}
}

func sendMessage(cln *utils.Client, message utils.Message) {
	cln.Dial.Write(utils.Enpack(message))
}

func receiveMessage(cln *utils.Client, timeout int) (utils.Message, error) {
	buffer := make([]byte, 1024)
	// cln.Dial.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	_, err := cln.Dial.Read(buffer)
	if err == nil {
		return utils.Depack(buffer), nil
	} else {
		return utils.Message{}, err
	}
}

func buildClientMessage(cln *utils.Client, action utils.Actions) utils.Message {
	return utils.Message{
		Action: action,
		Params: map[string]string{
			utils.TimestampField: time.Now().Format("2006-01-02 3:4:5 pm"),
			utils.TokenField:     cln.UserToken,
		},
	}
}

func outputMessage(cln *utils.Client, message utils.Message) {
	var out string
	switch message.Action {
	case utils.Srv:
		out = fmt.Sprintf("[%s] Server: %s\n", message.Params[utils.TimestampField], message.Params["text"])
	case utils.Mes:
		var author string
		if message.Params[utils.ReceiverTokenField] == cln.ReceiverToken {
			author = cln.ReceiverName
		} else {
			author = cln.User.Name
		}
		out = fmt.Sprintf("[%s] %s: %s", message.Params[utils.TimestampField], author, message.Params["text"])
	case utils.Err:
		out = fmt.Sprintf("[%s] Server (ERROR): %s\n", message.Params[utils.TimestampField], message.Params["text"])
	case utils.Ask:
		out = fmt.Sprintf("[%s] Server: %s\n", message.Params[utils.TimestampField], message.Params["text"])
	default:
		out = "Undefined action type...\n"
	}
	fmt.Print(out)
}
