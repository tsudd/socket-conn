package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

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
