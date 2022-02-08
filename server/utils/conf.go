package utils

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Addr  net.UDPAddr
	Users map[string]*User
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

func HandleConfig(path string) (Config, error) {
	configs := GetConfig(path)
	port, err := strconv.Atoi(GetElement("port", configs))
	if err != nil {
		return Config{}, err
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

	return Config{
		Addr: net.UDPAddr{
			IP:   net.ParseIP(GetElement("IP", configs)),
			Port: port,
			Zone: "",
		},
		// Users: map[string]*User{
		// 	"387e6278d8e06083d813358762e0ac63": {
		// 		Name: "joohncena",
		// 	},
		// },
		Users: users,
	}, nil

}
