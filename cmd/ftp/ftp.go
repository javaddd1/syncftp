package ftp //"SyncFTP/cmd/ftp"
import (
	"SyncFTP/cmd/config"
	"SyncFTP/cmd/ignore"
	"SyncFTP/cmd/logger"
	"SyncFTP/pkg/common"
	"fmt"
	"github.com/jlaffaye/ftp"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	URL         string
	PORT        string
	ID          string
	PWD         string
	FTP_PATH    string
	LOCAL_PATH  string
	IGNORE_PATH string
	DIRECTION   string
)

func init() {
	URL = config.LoadConfigInfomation("FTP", "HOST")
	PORT = config.LoadConfigInfomation("FTP", "PORT")
	ID = config.LoadConfigInfomation("FTP", "USER")
	PWD = config.LoadConfigInfomation("FTP", "PASSWORD")
	FTP_PATH = config.LoadConfigInfomation("FTP", "PATH")
	LOCAL_PATH = config.LoadConfigInfomation("LOCAL", "PATH")
	IGNORE_PATH = config.LoadConfigInfomation("IGNORE", "PATH")
	DIRECTION = strings.ToLower(config.LoadConfigInfomation("LOCAL", "DIRECTION"))
	//ファイル/ディレクトリ存在確認
	ExistIgnoreFile()
	ExistPath(LOCAL_PATH, FTP_PATH)
}

func Execute() {
	fmt.Println(common.INFO_MSG_01)
	if DIRECTION == "upload" || DIRECTION == "" {
		UploadSync()
	} else if DIRECTION == "download" {
		DownloadSync()
	}
	fmt.Println(common.INFO_MSG_02)
}

func FTPConnection() *ftp.ServerConn {
	client, err := ftp.Dial(URL + ":" + PORT)
	if err != nil {
		panic(err)
	}
	client.Login(ID, PWD)
	return client
}

func DownloadSync() {
	paths := FTPDirwalk(FTP_PATH)
	wait := sync.WaitGroup{}
	for _, path := range paths {
		wait.Add(1)
		go func(path string) {
			defer wait.Done()
			file_name := common.ExtFileName(path)
			//ignore対象の場合処理はしない
			if !ignore.IgnoreSync(path, file_name) {
				return
			}
			ext := strings.Replace(path, FTP_PATH, "", -1)
			local_path := LOCAL_PATH + ext
			byte_server_file, byte_local_file := GenerateByteData(path, local_path)
			//ファイルが異なる場合、既存ファイルのバックアップを取りダウンロード
			if common.Exists(local_path) {
				if string(byte_server_file) != string(byte_local_file) {
					CreateBackup(local_path)
					DownloadFile(local_path, byte_server_file)
				}
			} else {
				DownloadFile(local_path, byte_server_file)
			}
		}(path)
	}
	wait.Wait()
}

func UploadSync() {
	paths := Dirwalk(LOCAL_PATH)
	wait := sync.WaitGroup{}
	for _, path := range paths {
		wait.Add(1)
		go func(path string) {
			defer wait.Done()
			file_name := strings.Replace(path, filepath.Dir(path)+"\\", "", -1)
			path = strings.Replace(path, `\`, "/", -1)
			if !ignore.IgnoreSync(path, file_name) {
				return
			}
			//ループ外だと並行処理できないのでここで接続
			client := FTPConnection()
			defer client.Quit()
			ext := strings.Replace(path, LOCAL_PATH, "", -1)
			ftp_path := FTP_PATH + ext
			server_file, _ := client.Retr(ftp_path)
			if server_file == nil {
				err := FTPStor(client, ftp_path, path, file_name)
				if err != nil {
					panic(err)
				}
			} else {
				byte_server_file, byte_local_file := GenerateByteData(ftp_path, path)
				//nilでCloseするとエラーになる
				server_file.Close()
				//ファイルが異なる場合、既存ファイルのバックアップを取りアップロード
				if string(byte_server_file) != string(byte_local_file) {
					err := FTPCreateBackup(client, ftp_path, file_name)
					if err != nil {
						panic(err)
					}
					err = FTPStor(client, ftp_path, path, file_name)
					if err != nil {
						panic(err)
					}
				}
			}
		}(path)
	}
	//ゴルーチンが全て完了するまで待機
	wait.Wait()
}

func FTPStor(client *ftp.ServerConn, ftp_path string, path string, file_name string) error {
	MakeDir(ftp_path, FTP_PATH, file_name, 1)
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	err = client.Stor(ftp_path, f)
	fmt.Printf(common.INFO_MSG_03, file_name)
	logger.LogPrint(common.INFO_MSG_03, file_name)
	return err
}

func DownloadFile(path string, content []uint8) {
	file_name := common.ExtFileName(path)
	MakeDir(path, LOCAL_PATH, file_name, 0)
	ioutil.WriteFile(path, content, os.ModePerm)
	fmt.Printf(common.INFO_MSG_04, file_name)
	logger.LogPrint(common.INFO_MSG_04, file_name)
}

func FTPCreateBackup(client *ftp.ServerConn, path string, file_name string) error {
	err := client.Rename(path, path+"."+common.TimeNow())
	fmt.Printf(common.INFO_MSG_05, file_name)
	logger.LogPrint(common.INFO_MSG_05, file_name)
	return err
}

func CreateBackup(file_name string) {
	backup_name := file_name + "." + common.TimeNow()
	if err := os.Rename(file_name, backup_name); err != nil {
		panic(err)
	}
	file_name = common.ExtFileName(file_name)
	fmt.Printf(common.INFO_MSG_05, file_name)
	logger.LogPrint(common.INFO_MSG_05, file_name)
}

func Dirwalk(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, Dirwalk(filepath.Join(path, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(path, file.Name()))
	}
	return paths
}

func FTPDirwalk(path string) []string {
	client := FTPConnection()
	list, _ := client.List(path)
	var paths []string
	for _, l := range list {
		//対象がディレクトリの場合
		if l.Type == ftp.EntryTypeFolder {
			paths = append(paths, FTPDirwalk(path+"/"+l.Name)...)
			continue
		}
		paths = append(paths, path+"/"+l.Name)
	}
	return paths
}

func GenerateByteData(path string, local_path string) ([]uint8, []uint8) {
	client := FTPConnection()
	server_file, _ := client.Retr(path)
	byte_server_file, _ := ioutil.ReadAll(server_file)
	server_file.Close()

	local_file, _ := os.Open(local_path)
	byte_local_file, _ := ioutil.ReadAll(local_file)
	local_file.Close()

	return byte_server_file, byte_local_file
}

func ExistIgnoreFile() {
	if !common.Exists(IGNORE_PATH) {
		fmt.Println(common.ERROR_MSG_01)
		os.Exit(1)
	}
}

func MakeDir(path string, base_path string, file_name string, make_type int8) {
	dir_path := strings.Replace(path, file_name, "", -1)
	dir_path = strings.Replace(dir_path, base_path, "", -1)
	path_slice := common.Explode(dir_path, "/")
	make_path := base_path
	for _, p := range path_slice {
		if p != "" {
			make_path = make_path + "/" + p
			//ローカルにディレクトリ作成
			if make_type == 0 {
				if !common.Exists(make_path) {
					os.Mkdir(make_path, 0777)
				}
			//サーバにディレクトリ作成
			} else if make_type == 1 {
				client := FTPConnection()
				server_dir, _ := client.Retr(make_path)
				if server_dir == nil {
					client.MakeDir(make_path)
				}
			}
		}
	}
}

func ExistPath(local_path string, server_path string) {
	client := FTPConnection()
	_, err := client.List(server_path)
	if err != nil {
		fmt.Println(common.ERROR_MSG_02)
		os.Exit(1)
	}
	if f, err := os.Stat(local_path); os.IsNotExist(err) || !f.IsDir() {
		fmt.Println(common.ERROR_MSG_03)
		os.Exit(1)
	}
}
