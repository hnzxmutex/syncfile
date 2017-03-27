package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Client struct {
	conn    *xsocket
	path    string
	ignore  []*regexp.Regexp
	isWatch bool
	id      int64

	queueMap  map[string]int64
	queueLock *sync.Mutex
}

const DelaySync = 1

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
		conn:      xsocket,
		path:      path,
		ignore:    ignore,
		queueMap:  make(map[string]int64),
		queueLock: new(sync.Mutex),
	}
}

//给持续监控的时候用的队列
func (c *Client) pushFile(file string) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	//delay sync
	c.queueMap[file] = time.Now().Unix() + DelaySync
}

func (c *Client) popFile() []string {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	now := time.Now().Unix()
	var result []string
	for k, t := range c.queueMap {
		if t < now {
			delete(c.queueMap, k)
			result = append(result, k)
		}
	}
	return result
}

func (c *Client) SetWatch(isWatch bool) {
	c.isWatch = isWatch
}

func (c *Client) Start() {
	c.Sync(c.path)
	if c.isWatch {
		c.Watch(c.path)
	}
	c.conn.Close()
}

func (c *Client) Watch(watchPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println(event.Name, "change!")
				if _, err := os.Lstat(event.Name); !os.IsNotExist(err) {
					log.Println(event.Name, "push into queue!")
					c.pushFile(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	//add watcher
	filepath.Walk(watchPath, func(path string, f os.FileInfo, err error) error {
		relativePath, err := filepath.Rel(c.path, path)
		if c.isIgnore(relativePath) {
			return nil
		}
		if f.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
	c.processQueue()
	<-done
}

func (c *Client) processQueue() {
	ticker := time.NewTicker(time.Millisecond * DelaySync * 900)
	go func() {
		for {
			<-ticker.C
			file := c.popFile()
			if len(file) > 0 {
				for _, v := range file {
					log.Println("pop file:", v)
					c.Sync(v)
				}
			}
		}
	}()
}

func (c *Client) Sync(syncPath string) {
	filepath.Walk(syncPath, func(path string, f os.FileInfo, err error) error {
		c.id++
		if path == c.path {
			return nil
		}

		log.Println("=========walk file:", path, "================")
		//检查文件
		if c.checkFile(path) {
			log.Println("ok,send file")
			c.sendFile(path, f.Size())
			var result [3]byte
			_, err := c.conn.Read(result[:])
			if err != nil && err != io.EOF {
				log.Fatalln(err)
			}
			log.Println("send file over,server say:", string(result[:2]), "service id:", result[2])
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
	defer fileHandle.Close()
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
	if _, err := os.Lstat(src); os.IsNotExist(err) {
		log.Println(src, " not exist ignore")
		return false
	}
	file := getFileInfo(src)
	relativePath, err := filepath.Rel(c.path, src)
	if err != nil {
		log.Fatalln(err)
	}
	relativePath = strings.TrimLeft(relativePath, "..")
	//是否在白名单
	if c.isIgnore(relativePath) {
		return false
	}
	file.Name = relativePath
	cmdLine, _ := json.Marshal(file)

	length := len(cmdLine)
	var header [3]byte
	header[0] = byte(length >> 8)
	header[1] = byte(length & 0xff)
	header[2] = byte(c.id % 0xff)
	log.Println("check:", relativePath, "client id:", header[2])
	c.conn.Write(header[:])
	c.conn.Write(cmdLine)

	_, err = c.conn.Read(header[:])
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("server say:", string(header[:2]), "server id:", header[2])
	if header[0] == 'g' && header[1] == 'f' {
		return true
	} else if header[0] == 'i' && header[1] == 'g' {
		return false
	} else {
		log.Fatalln("error result:", string(header[:2]), "server id:", header[2])
		return false
	}
}
