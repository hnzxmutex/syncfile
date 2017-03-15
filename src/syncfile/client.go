package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Client struct {
	conn   *xsocket
	path   string
	ignore []*regexp.Regexp
}

func NewClient(serverAddr, path, password string, ignore []*regexp.Regexp) *Client {
	localAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Fatalln("addr error", err)
	}
	conn, err := net.DialTCP("tcp", nil, localAddr)
	if err != nil {
		log.Fatalln("connect error", err)
	}
	path, _ = filepath.Abs(path)
	xsocket := NewXSocket(conn, password)
	//握手
	var chat [4]byte
	xsocket.Write([]byte("ping")) //ping
	xsocket.Read(chat[:])         //pong
	if string(chat[:]) != "pong" {
		log.Fatalln("xsocket connect fail!")
		xsocket.Close()
		return nil
	}
	log.Println("xsocket connect success!")

	//握手完毕
	return &Client{
		conn:   xsocket,
		path:   path,
		ignore: ignore,
	}
}

func (c *Client) WatchAndSync() {
	c.SyncAll()
	c.conn.Close()
	return
}

func (c *Client) SyncAll() {
	filepath.Walk(c.path, func(path string, f os.FileInfo, err error) error {
		if path == c.path {
			return nil
		}

		log.Println("walk file:", path)
		//检查文件
		if c.checkFile(path) {
			log.Println("ok,send file")
			c.sendFile(path, f.Size())
			var result [2]byte
			_, err := c.conn.Read(result[:])
			if err != nil && err != io.EOF {
				log.Fatalln(err)
			}
			if result[0] == 'o' && result[1] == 'v' {
				log.Println("send file success:", path)
			} else if err == io.EOF {
				log.Fatalln("server close:)")
			} else {
				log.Fatalln("unknow result:", string(result[:]))
			}
		}
		return nil
	})
}

func (c *Client) sendFile(path string, size int64) {
	fileHandle, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("io.Copy")
	io.CopyN(c.conn, fileHandle, size)
}

func (c *Client) isIgnore(relativePath string) bool {
	for _, reg := range c.ignore {
		if reg.MatchString(relativePath) {
			log.Printf("[%s] match %s, ignore file", reg.String(), relativePath)
			return true
		}
	}
	return false
}

func (c *Client) checkFile(src string) bool {
	file := getFileInfo(src)
	relativePath, err := filepath.Rel(c.path, src)
	if err != nil {
		log.Fatalln(err)
	}
	relativePath = strings.TrimLeft(relativePath, ".")
	log.Println("check:", relativePath)
	//是否在白名单
	if c.isIgnore(relativePath) {
		return false
	}

	//connect
	cmdLine := []byte(fmt.Sprintf(
		"%s %d %d %d %s %t",
		relativePath,
		file.time.Unix(),
		file.perm,
		file.size,
		file.md5,
		file.isDir,
	))

	length := len(cmdLine)
	var header [2]byte
	header[0] = byte(length >> 8)
	header[1] = byte(length & 0xff)
	c.conn.Write(header[:])
	c.conn.Write(cmdLine)

	_, err = c.conn.Read(header[:])
	if err != nil {
		log.Fatalln(err)
		// } else {
		// 	log.Println("server return:", string(header[:]), ", n:", n)
	}
	return header[0] == 'g' && header[1] == 'f'
}
