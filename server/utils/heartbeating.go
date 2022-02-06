package utils

import (
	"net"
	"time"
)

func HeartBeating(conn net.UDPConn, readerChannel chan byte, timeout int) {
	select {
	case _ = <-readerChannel:
		LogMsg(conn.RemoteAddr().String(), "get message, keeping heartbeating...")
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		break
	case <-time.After(time.Second * 5):
		LogMsg("No messages in fact...")
		conn.Close()
	}

}

func GravelChannel(n []byte, mess chan byte) {
	for _, v := range n {
		mess <- v
	}
	close(mess)
}
