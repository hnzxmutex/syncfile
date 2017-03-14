package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	conn net.Conn
	path string
}

func NewClient(serverAddr, path string) *Client {
	localAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Fatalln("addr error", err)
	}
	conn, err := net.DialTCP("tcp", nil, localAddr)
	if err != nil {
		log.Fatalln("connect error", err)
	}
	path, _ = filepath.Abs(path)
	return &Client{
		conn: conn,
		path: path,
	}
}

func (c *Client) WatchAndSync() {
	filepath.Walk(c.path, func(path string, f os.FileInfo, err error) error {
		if path == c.path {
			return nil
		}

		log.Println("walk file:", path)
		//发送文件
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
	return
}

func (c *Client) sendFile(path string, size int64) {
	fileHandle, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("io.Copy")
	io.CopyN(c.conn, fileHandle, size)
}

func (c *Client) checkFile(src string) bool {
	file := getFileInfo(src)
	relativePath, err := filepath.Rel(c.path, src)
	//过滤..防止非法输入
	if err != nil {
		log.Fatalln(err)
	}
	relativePath = strings.TrimLeft(relativePath, ".")
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

	c.conn.Read(header[:])
	result := string(header[:])
	if result == "gf" {
		//get file
		return true
	} else if result == "if" {
		//ignore file
		return false
	} else {
		log.Fatalln("unknow result", result)
		return false
	}
}

func (c *Client) Handler() {

}
