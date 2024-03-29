package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
)

const (
	MessageLengthNumber = 4
	TokenField          = "token"
	SenderField         = "sender"
	ReceiverTokenField  = "reser"
	TimestampField      = "time"
	SenderNameField     = "myname"
	TextField           = "text"
)

type Actions int

const (
	Con Actions = iota
	Mes
	Syn
	Ask
	Srv
	Err
	Hrt
)

type Message struct {
	Action Actions
	Params map[string]string
}

func Enpack(message Message) []byte {
	ret, _ := json.Marshal(message)
	return append(IntToBytes(len(ret)), ret...)
}

func Depack(buffer []byte) Message {
	length := len(buffer)

	var i int
	data := make([]byte, 1024)
	for i = 0; i < length; i++ {
		if length < i+MessageLengthNumber {
			break
		}
		messageLen := BytesToInt(buffer[i : i+MessageLengthNumber])
		if length < i+MessageLengthNumber+messageLen {
			break
		}
		data = buffer[i+MessageLengthNumber : i+MessageLengthNumber+messageLen]
	}
	if i == length {
		return Message{}
	}
	var res Message
	err := json.Unmarshal(data, &res)
	if err != nil {
		LogErr(err)
	}
	return res
}

func IntToBytes(n int) []byte {
	x := int32(n)

	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, x)
	return buffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
