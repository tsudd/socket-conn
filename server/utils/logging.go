package utils

import (
	"log"
	"os"
)

func LogErr(v ...interface{}) {

	logfile := os.Stdout
	log.Println(v...)
	logger := log.New(logfile, "\r\n", log.Llongfile|log.Ldate|log.Ltime)
	logger.SetPrefix("[Error]")
	logger.Println(v...)
	defer logfile.Close()
}

func LogMsg(v ...interface{}) {

	logfile := os.Stdout
	log.Println(v...)
	logger := log.New(logfile, "\r\n", log.Ldate|log.Ltime)
	logger.SetPrefix("[Info]")
	logger.Println(v...)
	defer logfile.Close()
}

func ChkErr(err error) {
	if err != nil {
		LogErr(os.Stderr, "Fatal error: %s", err.Error())
	}
}
