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
