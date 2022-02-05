package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/tsudd/socket-conn/server/utils"
)

type Msg struct {
	Meta    map[string]interface{} `json:"meta"`
	Content interface{}            `json:"content"`
}

func send(conn net.Conn) {
	for i := 0; i < 6; i++ {
		session := GetSession()
		message := &Msg{
			Meta: map[string]interface{}{
				"meta": "test",
				"ID":   strconv.Itoa(i),
			},
			Content: Msg{
				Meta: map[string]interface{}{
					"author": "nucky lu",
				},
				Content: session,
			},
		}
		result, _ := json.Marshal(message)
		conn.Write(utils.Enpack(result))
		time.Sleep(1 * time.Second)
	}
	fmt.Println("send over")
	defer conn.Close()
}

func GetSession() string {
	gs1 := time.Now().Unix()
	gs2 := strconv.FormatInt(gs1, 10)
	return gs2
}

func main() {
	buffer := make([]byte, 2048)
	server := "localhost:2333"
	udpAddr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	conn, err := net.Dial("udp4", udpAddr.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	fmt.Println("connect success")
	fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
	_, err = bufio.NewReader(conn).Read(buffer)
	if err == nil {
		fmt.Printf("%s\n", buffer)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
	// send(conn)

}
