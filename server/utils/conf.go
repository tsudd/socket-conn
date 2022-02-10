package utils

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type SConfig struct {
	Addr  net.UDPAddr
	Users map[string]*User
}

type CConfig struct {
	Addr      net.UDPAddr
	UserToken string
}

func GetConfig(path string) map[interface{}]interface{} {
	data, err := os.ReadFile(path)
	m := make(map[interface{}]interface{})
	if err != nil {
		LogErr(err)
	}
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		LogErr(err)
	}
	return m
}

func GetElement(key string, mapping map[interface{}]interface{}) string {
	if value, ok := mapping[key]; ok {
		return fmt.Sprint(value)
	}
	LogMsg("no config")
	return ""
}

func HandleServerConfig(path string) (SConfig, error) {
	configs := GetConfig(path)
	port, err := strconv.Atoi(GetElement("port", configs))
	if err != nil {
		return SConfig{}, err
	}
	users := make(map[string]*User)
	configUsers, ok := configs["users"]
	if ok {
		for token, user := range configUsers.(map[interface{}]interface{}) {
			userFields := user.(map[interface{}]interface{})
			users[token.(string)] = &User{
				Name: userFields["name"].(string),
			}
		}
	} else {
		LogErr("Couldn't parse users", err)
	}

	return SConfig{
		Addr: net.UDPAddr{
			IP:   net.ParseIP(GetElement("IP", configs)),
			Port: port,
			Zone: "",
		},
		Users: users,
	}, nil
}

func HandleClientConfig(path string) (CConfig, error) {
	configs := GetConfig(path)
	port, err := strconv.Atoi(GetElement("port", configs))
	if err != nil {
		return CConfig{}, err
	}
	host := GetElement("host", configs) + ":" + fmt.Sprint(port)
	udpAddr, err := net.ResolveUDPAddr("udp4", host)
	if err != nil {
		return CConfig{}, err
	}
	return CConfig{
		Addr:      *udpAddr,
		UserToken: GetElement("user", configs),
	}, nil
}
