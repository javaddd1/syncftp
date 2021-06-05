package config //"SyncFTP/cmd/config"

import (
	"SyncFTP/pkg/common"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"strings"
)

func ReadConfig() *ini.File {
	flag.Parse()
	path := strings.Join(flag.Args(), "")
	if path == "" {
		fmt.Println(common.ERROR_MSG_02)
		os.Exit(1)
	}
	if !common.Exists(path) {
		fmt.Println(common.ERROR_MSG_03)
		os.Exit(1)
	}
	Config, _ := ini.Load(path)
	return Config
}

func LoadConfigInfomation(section string, key string) string {
	Config := ReadConfig()
	return Config.Section(section).Key(key).String()
}
