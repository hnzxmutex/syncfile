package cmd

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	LOG_BLACK    = "30"
	LOG_RAD      = "31"
	LOG_GREEN    = "32"
	LOG_YELLO    = "33"
	LOG_BLUE     = "34"
	LOG_PURPLE   = "35"
	LOG_SKY_BLUE = "36"
	LOG_WHITE    = "37"
)

type CommandType byte

const (
	_                   CommandType = iota
	COMMAND_GET_FILE                //请客户端发送文件
	COMMAND_SEND_OVER               //发送完毕
	COMMAND_IGNORE_FILE             //忽略文件不要发送
)

const (
	PING = "123%^2sync_ping"
	PONG = "345$#1sync_pong"
)

type SysFileInfo struct {
	Name  string
	Size  int64
	Time  time.Time
	Perm  os.FileMode
	Md5   string
	IsDir bool
}

func getFileInfo(filePath string) *SysFileInfo {
	fi, err := os.Lstat(filePath)
	if err != nil {
		log.Fatalln("info error:", err)
	}
	fileHandle, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("open error:", err)
	}
	defer fileHandle.Close()

	h := md5.New()
	_, err = io.Copy(h, fileHandle)

	return &SysFileInfo{
		Name:  fi.Name(),
		Size:  fi.Size(),
		Perm:  fi.Mode().Perm(),
		Time:  fi.ModTime(),
		Md5:   fmt.Sprintf("%x", h.Sum(nil)),
		IsDir: fi.IsDir(),
	}
}

func trimDoubleDot(str string) string {
	if len(str) < 2 {
		return str
	}
	strByte := []byte(str)
	for len(strByte) >= 2 {
		if !(strByte[0] == '.' && strByte[1] == '.') {
			break
		}
		strByte = strByte[2:]
	}
	return string(strByte)
}

func highlightLog(str, color string) string {
	return "\033[" + color + "m" + str + "\033[0m"
}

//ignore list
func getIgnoreList(path string) []*regexp.Regexp {
	ignoreData, err := ioutil.ReadFile(path)
	var ignoreListString []string
	if err != nil {
		log.Println("Warning:read ignore file error,sync all file")
		return []*regexp.Regexp{}
	} else {
		ignoreString := strings.TrimLeft(string(ignoreData), " ")
		ignoreString = strings.Replace(ignoreString, "\r\n", "\n", -1)
		ignoreString = strings.Replace(ignoreString, "\r", "\n", -1)
		ignoreString = strings.Trim(string(ignoreData), "\n")
		ignoreListString = strings.Split(ignoreString, "\n")
		if len(ignoreListString) == 1 && ignoreListString[0] == "" {
			ignoreListString = []string{}
		}
	}

	regList := make([]*regexp.Regexp, 0, len(ignoreListString))
	for _, p := range ignoreListString {
		p = strings.TrimLeft(p, " ")
		if p == "" || ([]rune(p))[0] == '#' {
			continue
		}
		regList = append(regList, regexp.MustCompile(p))
	}
	return regList
}
