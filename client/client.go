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
	OnlineCommand    = "online"
)

const (
	HelpMessage = "Help: commands:\nchat <username> - start chatting with specific user, user should be connected to the server\n" +
		"online - get a list of connected to the server users\nexit - end client programm\nhelp - get help\n"
	Heartbeating = 7
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
		User:       &utils.User{Name: settings.Username},
		ServerAddr: settings.Addr,
		Mode:       utils.Wait,
	}

	fmt.Println("Init of UDP client with ", cln.User.Name, " user for server ", cln.ServerAddr.IP.String(), ":", cln.ServerAddr.Port)
	fmt.Printf("Starting connection to %s:%d....\n", cln.ServerAddr.IP.String(), cln.ServerAddr.Port)
	err = getConnection(cln)
	if err != nil {
		utils.ChkErr(err)
		os.Exit(0)
	}
	fmt.Println("Succesfully connected to the server!")
	go handleReading(cln)
	go startHeartBeating(cln)
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
			case OnlineCommand:
				mes := buildClientMessage(cln, utils.Con)
				sendMessage(cln, mes)
			case HelpCommand:
				fmt.Print(HelpMessage)
			default:
				fmt.Printf("Unknown command...\n")
			}
		} else {
			mes := buildClientMessage(cln, utils.Mes)
			mes.Params[utils.ReceiverTokenField] = cln.ReceiverToken
			mes.Params[utils.TextField] = input
			cln.History = append(cln.History, fmt.Sprintf("[%s] %s: %s", mes.Params[utils.TimestampField], mes.Params[cln.User.Name], input))
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
		case utils.Syn:
			answer := buildClientMessage(cln, utils.Syn)
			answer.Params[utils.TextField] = strings.Join(cln.History, "")
			answer.Params[utils.SenderNameField] = cln.User.Name
			answer.Params[utils.ReceiverTokenField] = mes.Params[utils.TokenField]
			cln.ReceiverName = mes.Params[utils.SenderNameField]
			cln.Mode = utils.Chat
			cln.ReceiverToken = mes.Params[utils.TokenField]
			go sendMessage(cln, answer)
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
	connRequest.Params[utils.TokenField] = cln.User.Name
	sendMessage(cln, connRequest)
	mes, err := receiveMessage(cln, 5)
	if err != nil {
		return err
	}
	if mes.Action == utils.Con {
		cln.UserToken = mes.Params[utils.SenderField]
		fmt.Print(mes.Params[utils.TextField])
		return nil
	} else {
		return errors.New(mes.Params[utils.TextField])
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
		out = fmt.Sprintf("[%s] Server: %s\n", message.Params[utils.TimestampField], message.Params[utils.TextField])
	case utils.Syn:
		out = fmt.Sprintf(message.Params[utils.TextField])
	case utils.Mes:
		var author string
		if message.Params[utils.TokenField] == cln.UserToken {
			author = cln.ReceiverName
		} else {
			author = cln.User.Name
		}
		out = fmt.Sprintf("[%s] %s: %s", message.Params[utils.TimestampField], author, message.Params[utils.TextField])
		cln.History = append(cln.History, out)
	case utils.Err:
		out = fmt.Sprintf("[%s] Server (ERROR): %s\n", message.Params[utils.TimestampField], message.Params[utils.TextField])
	case utils.Ask:
		out = fmt.Sprintf("[%s] Server: %s\n", message.Params[utils.TimestampField], message.Params[utils.TextField])
	default:
		out = "Undefined action type...\n"
	}
	fmt.Print(out)
}

func startHeartBeating(cln *utils.Client) {
	for {
		heart := buildClientMessage(cln, utils.Hrt)
		go sendMessage(cln, heart)
		time.Sleep(time.Duration(Heartbeating) * time.Second)
	}
}
