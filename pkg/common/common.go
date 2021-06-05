package common //"SyncFTP/pkg/common"

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func Explode(tPath string, sep string) []string {
	var ret []string
	for _, s := range strings.Split(tPath, sep) {
		if s != "" {
			ret = append(ret, s)
		}
	}
	return ret
}

func IsMatch(data string, file_name string) (bool, error) {
	return path.Match(data, file_name)
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func TimeNow() string {
	day := time.Now()
	const layout = "20060102150405"
	return day.Format(layout)
}

func ExtFileName(path string) string {
	return filepath.Base(path)
}
