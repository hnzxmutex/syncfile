package main

import (
	"io"
	"net"
)

type xsocket struct {
	net.Conn
	password []byte
	pl       int
}

func NewXSocket(conn net.Conn, password string) *xsocket {
	return &xsocket{
		Conn:     conn,
		password: []byte(password),
		pl:       len(password),
	}
}

func (s *xsocket) Write(data []byte) (n int, err error) {
	for i := 0; i < len(data); i++ {
		data[i] = data[i] ^ s.password[i%s.pl]
	}
	x := 0
	all := len(data)

	for all > 0 {
		n, err = s.Conn.Write(data)
		if err != nil {
			n += x
			return
		}
		all -= n
		x += n
		data = data[n:]
	}
	return
}

func (s *xsocket) Read(data []byte) (n int, err error) {
	n, err = io.ReadFull(s.Conn, data)
	//简单解密
	if n > 0 {
		for i := 0; i < n; i++ {
			data[i] = data[i] ^ s.password[i%s.pl]
		}
	}

	return
}
