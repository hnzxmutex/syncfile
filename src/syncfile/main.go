package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type SysFileInfo struct {
	name string
	size int64
	time time.Time
	perm os.FileMode
	md5  string
}

func main() {
	clientSend("Makefile")
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

	h := md5.New()
	_, err = io.Copy(h, fileHandle)

	return &SysFileInfo{
		name: fi.Name(),
		size: fi.Size(),
		perm: fi.Mode().Perm(),
		time: fi.ModTime(),
		md5:  fmt.Sprintf("%x", h.Sum(nil)),
	}
}

func clientSend(src string) {
	file := getFileInfo(src)
	log.Println(file)
}