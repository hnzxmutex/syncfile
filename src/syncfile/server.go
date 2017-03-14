package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	l    net.Listener
	path string
}

func NewServer(addr, path string) *Server {

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
		l:    l,
		path: path,
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
		go s.Handler(conn)
	}
}

func (s *Server) Handler(conn net.Conn) {
	log.Println("new client is connected")
	defer conn.Close()
	for {
		log.Println("===========check a new file===============")
		fi, err := s.getFileInfo(conn)
		if err != nil {
			log.Println(err)
			return
		}
		fi.name = filepath.Join(s.path, strings.Replace(fi.name, "..", ".", -1))
		log.Println(fi.name)
		if fi.isDir {
			//文件夹
			log.Println("create dir:", fi.name, fi.perm)
			if err := os.MkdirAll(fi.name, fi.perm); err != nil {
				log.Println(err)
			}
		} else if _, err := os.Lstat(fi.name); os.IsNotExist(err) {
			fileHandle, err := s.createFile(fi)
			if err != nil {
				log.Println(err)
			} else {
				//是文件,且文件不存在
				if fi.size != 0 {
					conn.Write([]byte("gf")) //get file
					s.saveFile(conn, fileHandle, fi.size)
					fileHandle.Close()
					continue
				}
				fileHandle.Close()
			}
		} else {
			log.Println(fi.name, "file exists check md5")
			localfi := getFileInfo(fi.name)
			if localfi.md5 != fi.md5 {
				fileHandle, err := os.Create(fi.name)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("file md5 different")
					s.saveFile(conn, fileHandle, fi.size)
					fileHandle.Close()
					continue
				}
			}
		}
		conn.Write([]byte("if")) //ignore file
	}
}

func (s *Server) createFile(fi *SysFileInfo) (*os.File, error) {
	log.Println("create file:", fi.name)
	if err := os.MkdirAll(filepath.Dir(fi.name), fi.perm); err != nil {
		return nil, err
	}
	//open file
	fileHandle, err := os.OpenFile(fi.name, os.O_RDWR|os.O_CREATE, fi.perm)
	if err != nil {
		return nil, err
	}
	return fileHandle, nil
}

func (s *Server) saveFile(conn net.Conn, fileHandle *os.File, size int64) {
	conn.Write([]byte("gf")) //get file
	log.Println("writing file")
	num, err := io.CopyN(fileHandle, conn, size)
	if err != nil {
		log.Println("write file err:", err)
	} else {
		log.Println("write success,size:", num)
	}
	conn.Write([]byte("ov")) //ov
}

func (s *Server) getFileInfo(conn net.Conn) (*SysFileInfo, error) {
	//解析头部
	var header [2]byte
	_, err := conn.Read(header[:])
	if err != nil {
		return nil, err
	}
	length := int(header[0])<<8 | int(header[1])
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
	var fileName, md5 string
	var filePerm, fileMtime, fileSize int64
	var isDir bool

	n, err := fmt.Sscanf(
		string(str),
		"%s %d %d %d %s %t",
		&fileName,
		&fileMtime,
		&filePerm,
		&fileSize,
		&md5,
		&isDir,
	)

	if n != 6 {
		return nil, errors.New("invalid header")
	} else if err != nil {
		return nil, err
	}

	return &SysFileInfo{
		name:  fileName,
		time:  time.Unix(fileMtime, 0),
		perm:  os.FileMode(filePerm),
		size:  fileSize,
		md5:   md5,
		isDir: isDir,
	}, nil
}
