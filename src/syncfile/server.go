package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Server struct {
	l        net.Listener
	path     string
	ignore   []*regexp.Regexp
	password string
}

var i int64 = 0

func NewServer(addr, path, password string, ignore []*regexp.Regexp) *Server {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("net failed", err)
	}
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Start Service")
	path, _ = filepath.Abs(path)
	return &Server{
		l:        l,
		path:     path,
		ignore:   ignore,
		password: password,
	}
}

func (s *Server) Serve() {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			log.Println("network error:", err)
		}

		xsocket := NewXSocket(conn, s.password)
		//握手
		var chat [4]byte
		xsocket.Read(chat[:])
		if string(chat[:]) == "ping" {
			xsocket.Write([]byte("pong"))
		} else {
			xsocket.Close()
			log.Println("xsocket connect fali!")
			continue
		}
		log.Println("xsocket connect success!")
		//握手完毕
		go s.Handler(xsocket)
	}
}

func (s *Server) isIgnore(relativePath string) bool {
	for _, reg := range s.ignore {
		if reg.MatchString(relativePath) {
			log.Printf("[%s] match %s, ignore file", reg.String(), relativePath)
			return true
		}
	}
	return false
}

func (s *Server) Handler(conn *xsocket) {
	log.Println("new client is connected")
	defer conn.Close()
	for {
		fi, err := s.getFileInfo(conn)
		i++
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
		if s.isIgnore(fi.Name) {
			s.ignoreFile(conn)
			continue
		}
		//绝对路径
		fi.Name = filepath.Join(s.path, fi.Name)
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
	conn.Write([]byte{'g', 'f', byte(i % 0xff)}) //get file
	log.Println("writing file, server id:", byte(i%0xff))
	num, err := io.CopyN(fileHandle, conn, size)
	if err != nil {
		log.Println("write file err:", err, "server id:", byte(i%0xff))
	} else {
		log.Println("write success,size:", num, "server id:", byte(i%0xff))
	}
	conn.Write([]byte{'o', 'v', byte(i % 0xff)}) //get file
}

func (s *Server) ignoreFile(conn *xsocket) {
	log.Println("ignore ,send ov;server id:", byte(i%0xff))
	conn.Write([]byte([]byte{'i', 'g', byte(i % 0xff)})) //ignore file
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
