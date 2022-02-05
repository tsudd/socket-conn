package utils

import (
	"bytes"
	"encoding/binary"
)

const (
	MessageLengthNumber = 4
)

func Enpack(message []byte) []byte {
	return append(IntToBytes(len(message)), message...)
}

func Depack(buffer []byte) []byte {
	length := len(buffer)

	var i int
	data := make([]byte, 32)
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
		return make([]byte, 0)
	}
	return data
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
