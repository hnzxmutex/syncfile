package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	serverConfigPath string
)

type SyncFold struct {
	password, path, ignoreConfigFile string
	ignoreList                       []*regexp.Regexp
}

func (s *SyncFold) isIgnore(relativePath string) bool {
	for _, reg := range s.ignoreList {
		if reg.MatchString(relativePath) {
			log.Printf("[%s] match %s, ignore file", reg.String(), relativePath)
			return true
		}
	}
	return false
}

type Server struct {
	l           net.Listener
	id          uint64
	syncObjList []SyncFold
	lock        *sync.RWMutex
}

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "服务端模式",
	Long:  "服务端模式,接受客户端上传的文件",
	//Run: func(cmd *cobra.Command, args []string) {
}

func init() {
	serverCommand.Flags().StringVarP(&serverConfigPath, "config", "c", "", "服务端配置文件")
	serverCommand.Run = serverRun
}

func serverRun(cmd *cobra.Command, args []string) {
	viper.SetConfigFile(serverConfigPath)
	if serverConfigPath == "" {
		fmt.Println("example:./bin/syncfile -c ./server.yaml")
		cmd.Help()
		return
	}
	viper.SetConfigType("yaml")
	port := viper.GetString("port")
	syncObjList := loadYamlConfig()
	s := NewServer(port, syncObjList)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.syncObjList = loadYamlConfig()
	})
	s.Serve()
}

func loadYamlConfig() []SyncFold {
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("读取配置文件失败:", err)
	}
	var syncObjList []SyncFold

	for _, key := range viper.GetStringSlice("app_list") {
		currentConfig := viper.GetStringMapString(key)
		syncPath, pathExist := currentConfig["path"]
		password, passExist := currentConfig["password"]
		ignoreConfigFile := currentConfig["ignore_config_file"]
		if pathExist && passExist {
			if !filepath.IsAbs(syncPath) {
				log.Println("path must absolute:", syncPath)
				continue
			}
			err = os.MkdirAll(syncPath, 0755)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Start Service")
			syncPath, _ = filepath.Abs(syncPath)

			syncObjList = append(syncObjList, SyncFold{
				password:   password,
				path:       syncPath,
				ignoreList: getIgnoreList(ignoreConfigFile),
			})
		}
	}
	if len(syncObjList) == 0 {
		log.Fatalln("Config file error")
	}
	return syncObjList

}

func NewServer(port string, syncObjList []SyncFold) *Server {
	addr := fmt.Sprintf(":%s", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("net failed", err)
	}

	//为了有权限在端口listen,下面才setuid
	// if uid == 0 {
	// 	log.Fatalln("must set uid, run as root is not safe")
	// }
	// err = syscall.Setuid(uid)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	return &Server{
		l:           l,
		lock:        new(sync.RWMutex),
		syncObjList: syncObjList,
	}
}

func (s *Server) Serve() {
accept:
	for {
		conn, err := s.l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			log.Println("network error:", err)
		}
		//check password
		var chat, copyChar [len(PING)]byte
		n, err := io.ReadFull(conn, chat[:])
		if n == len(PING) && err == nil {
			s.lock.RLock()
			for _, syncObj := range s.syncObjList {
				//握手成功
				log.Println("handshake password:", syncObj.password)
				copy(copyChar[:], chat[:])
				if s.handShake(copyChar[:], []byte(syncObj.password)) {
					xsocket := NewXSocket(conn, syncObj.password)
					xsocket.Write([]byte(PONG))
					//握手完毕
					log.Println("xsocket connect success!")
					s.lock.RUnlock()
					go s.Handler(xsocket, syncObj)
					continue accept
				}
			}
			s.lock.RUnlock()
		}
		//握手失败
		log.Println("handshake fail")
		conn.Close()
		continue

	}
}

func (s *Server) handShake(pingString, password []byte) bool {
	pl := len(password)
	for i := 0; i < len(pingString); i++ {
		pingString[i] = pingString[i] ^ password[i%pl]
	}
	return string(pingString) == PING
}

func (s *Server) Handler(conn *xsocket, syncObj SyncFold) {
	log.Println("new client is connected")
	defer conn.Close()
	for {
		fi, err := s.getFileInfo(conn)
		s.id++
		if err == io.EOF {
			log.Println("connection close")
			return
		} else if err != nil {
			log.Println(err)
			log.Println("connection close")
			return
		}
		log.Println("===========check a new file===============")
		//过滤不安全字符
		//windows server
		if runtime.GOOS == "windows" {
			fi.Name = strings.Replace(fi.Name, "/", "\\", -1)
		} else {
			fi.Name = strings.Replace(fi.Name, "\\", "/", -1)
		}
		fi.Name = strings.Replace(fi.Name, "..", ".", -1)
		if syncObj.isIgnore(fi.Name) {
			s.ignoreFile(conn)
			continue
		}
		//绝对路径
		fi.Name = filepath.Join(syncObj.path, fi.Name)
		log.Println("get a file:", fi.Name)
		if fi.IsDir {
			//文件夹
			log.Println("create dir:", fi.Name, fi.Perm)
			if err := os.MkdirAll(fi.Name, fi.Perm); err != nil {
				log.Println(err)
			}
			os.Chmod(fi.Name, fi.Perm)
			s.ignoreFile(conn)
			continue
		}

		//文件不存在
		if _, err := os.Lstat(fi.Name); os.IsNotExist(err) {
			log.Println(fi.Name, "not found, create")
			fileHandle, err := s.createFile(fi)
			if err != nil {
				log.Println(err)
			} else {
				//是文件,且文件不存在
				if fi.Size != 0 {
					s.saveFile(conn, fileHandle, fi.Size)
					fileHandle.Close()
					continue
				}
				fileHandle.Close()
			}
		} else {
			//文件存在
			log.Println(fi.Name, "exist, check md5")
			localfi := getFileInfo(fi.Name)
			//md5不一致,同步
			if localfi.Md5 != fi.Md5 {
				fileHandle, err := os.Create(fi.Name)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("md5 of files different")
					s.saveFile(conn, fileHandle, fi.Size)
					fileHandle.Close()
					continue
				}
			} else {
				log.Println("md5 of files equal")
			}
		}
		s.ignoreFile(conn)
	}
}

func (s *Server) createFile(fi *SysFileInfo) (*os.File, error) {
	log.Println("create file:", fi.Name)
	if err := os.MkdirAll(filepath.Dir(fi.Name), fi.Perm); err != nil {
		return nil, err
	}
	//open file
	fileHandle, err := os.OpenFile(fi.Name, os.O_RDWR|os.O_CREATE, fi.Perm)
	if err != nil {
		return nil, err
	}
	return fileHandle, nil
}

func (s *Server) saveFile(conn *xsocket, fileHandle *os.File, size int64) {
	conn.Write([]byte{'g', 'f', byte(s.id % 0xff)}) //get file
	log.Println("writing file, server id:", byte(s.id%0xff))
	num, err := io.CopyN(fileHandle, conn, size)
	if err != nil {
		log.Println("write file err:", err, "server id:", byte(s.id%0xff))
	} else {
		log.Println("write success,size:", num, "server id:", byte(s.id%0xff))
	}
	conn.Write([]byte{'o', 'v', byte(s.id % 0xff)}) //get file
}

func (s *Server) ignoreFile(conn *xsocket) {
	log.Println("ignore ,send ov;server id:", byte(s.id%0xff))
	conn.Write([]byte([]byte{'i', 'g', byte(s.id % 0xff)})) //ignore file
}

func (s *Server) getFileInfo(conn *xsocket) (*SysFileInfo, error) {
	//解析头部
	var header [3]byte
	_, err := conn.Read(header[:])
	if err != nil {
		return nil, err
	}
	length := int(header[0])<<8 | int(header[1])
	log.Println("header size:", length, "client id:", header[2])
	if length == 0 || length > 0xfff {
		log.Println(length)
		return nil, errors.New("length not valid")
	}
	buf := make([]byte, length)
	_, err = conn.Read(buf[:])
	if err != nil {
		return nil, err
	}
	return s.parseHeader(buf)
}

func (s *Server) parseHeader(str []byte) (*SysFileInfo, error) {
	var fi SysFileInfo

	if err := json.Unmarshal(str, &fi); err != nil {
		log.Println("json解析失败", err)
		return nil, errors.New("微信api json解析失败")
	}
	return &fi, nil
}
