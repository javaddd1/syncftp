package ignore //"SyncFTP/cmd/ignore"

import (
	"SyncFTP/cmd/config"
	"SyncFTP/pkg/common"
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var (
	LOCAL_PATH  string
	FTP_PATH    string
	IGNORE_PATH string
	DIRECTION   string
)

func init() {
	FTP_PATH = config.LoadConfigInfomation("FTP", "PATH")
	LOCAL_PATH = config.LoadConfigInfomation("LOCAL", "PATH")
	IGNORE_PATH = config.LoadConfigInfomation("IGNORE", "PATH")
	DIRECTION = config.LoadConfigInfomation("LOCAL", "DIRECTION")
}

func LoadIgnoreData() []string {
	ignore_data, _ := os.Open(IGNORE_PATH)
	defer ignore_data.Close()
	scanner := bufio.NewScanner(ignore_data)

	data_array := []string{}
	for scanner.Scan() {
		data_array = append(data_array, scanner.Text())
	}
	return data_array
}

func IgnoreSync(tPath string, file_name string) bool {
	ext := ""
	sync_flag := true
	tPath = filepath.Dir(tPath)
	if DIRECTION == "upload" || DIRECTION == "" {
		ext = strings.Replace(LOCAL_PATH, FTP_PATH, "", -1)
		ext = strings.Replace(ext, "/", `\`, -1)
		tPath = strings.Replace(tPath, ext, "", -1)
	}
	for _, data := range LoadIgnoreData() {
		data = strings.Replace(data, "/", `\`, -1)
		if sync_flag {
			//対象がディレクトリの場合
			if strings.Contains(data, `\`) {
				tpath_slice := common.Explode(tPath, `\`)
				data_slice := common.Explode(data, `\`)
				if len(tpath_slice) < len(data_slice) {
					continue
				}
				for i := 0; i <= len(data_slice)-1; i++ {
					if strings.Contains(data_slice[i], "*") {
						if isMatch, _ := common.IsMatch(data_slice[i], tpath_slice[i]); isMatch {
							sync_flag = false
						} else {
							sync_flag = true
							break
						}
					} else {
						if data_slice[i] == tpath_slice[i] {
							sync_flag = false
						} else {
							sync_flag = true
							break
						}
					}
				}
				if len(data_slice) == 1 {
					for i := 0; i <= len(tpath_slice)-1; i++ {
						if isMatch, _ := common.IsMatch(data_slice[0], tpath_slice[i]); isMatch {
							sync_flag = false
							break
						}
					}
				}
			} else {
				if strings.Contains(data, "*") {
					if isMatch, _ := common.IsMatch(data, file_name); isMatch {
						sync_flag = false
					}
				} else {
					if data == file_name {
						sync_flag = false
					}
				}
			}
		}
	}
	return sync_flag
}
