package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/tsudd/socket-conn/server/utils"
)

func send(conn net.Conn) {
	message := utils.Message{
		Action: utils.Mes,
		Params: map[string]string{
			"nice":           "test",
			utils.TokenField: "387e6278d8e06083d813358762e0ac63",
			"text":           "Hi UDP Server, How are you doing?",
		},
	}
	conn.Write(utils.Enpack(message))
	time.Sleep(1 * time.Second)
	fmt.Println("send over")
}

func GetSession() string {
	gs1 := time.Now().Unix()
	gs2 := strconv.FormatInt(gs1, 10)
	return gs2
}

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
	// buffer := make([]byte, 2048)
	// server := "127.0.0.1:2333"
	// udpAddr, err := net.ResolveUDPAddr("udp4", server)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	// 	os.Exit(1)
	// }
	// conn, err := net.DialUDP("udp4", nil, udpAddr)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	// 	os.Exit(1)
	// }
	// defer conn.Close()
	// fmt.Println("connect success")
	// send(conn)
	// conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	// _, err = conn.Read(buffer)
	// if err == nil {
	// 	fmt.Printf("%s\n", buffer)
	// } else {
	// 	fmt.Printf("Some error %v\n", err)
	// }
	// // send(conn)

}

func startClient(path string) {
	fmt.Printf("Starting client using config from %s...\n", path)
	settings, err := utils.HandleClientConfig(path)
	utils.ChkErr(err)

	cln := &utils.Client{
		UserToken:  settings.UserToken,
		ServerAddr: settings.Addr,
	}

	fmt.Println("Init of UDP client with ", cln.UserToken, " token for server ", cln.ServerAddr.IP.String(), ":", cln.ServerAddr.Port)
	fmt.Printf("Starting connection to %s:%d....\n", cln.ServerAddr.IP.String(), cln.ServerAddr.Port)
	err = getConnection(cln)
	if err != nil {
		utils.ChkErr(err)
	}
	fmt.Println("Succesfully connected to the server!")

}

func handleReading(cln *utils.Client) {
	fmt.Println("Starting reading from server...")
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
	cln.Dial.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
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
			utils.TimestampField: time.Now().String(),
			utils.TokenField:     cln.UserToken,
		},
	}
}
