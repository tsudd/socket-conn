package main

import (
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
	buffer := make([]byte, 2048)
	server := "127.0.0.1:2333"
	udpAddr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("connect success")
	send(conn)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, err = conn.Read(buffer)
	if err == nil {
		fmt.Printf("%s\n", buffer)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	// send(conn)

}
