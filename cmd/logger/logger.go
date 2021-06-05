package logger //"SyncFTP/cmd/logger"

import (
	"SyncFTP/cmd/config"
	"SyncFTP/pkg/common"
	"log"
	"os"
)

var (
	LOG_PATH string
)

func init() {
	LOG_PATH = config.LoadConfigInfomation("LOG", "PATH")
}

func OpenLogFile() *os.File {
	common.Exists(LOG_PATH)
	log_file, _ := os.OpenFile(LOG_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	return log_file
}

func LogInit() {
	log_file := OpenLogFile()
	log.SetOutput(log_file)
	log.SetFlags(log.Ldate | log.Ltime)
}

func LogPrint(str string, file_name string) {
	LogInit()
	log.Printf("[INFO]"+str, file_name)
}
