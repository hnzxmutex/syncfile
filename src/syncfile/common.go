package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
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
