package main

import (
	"flag"
	"fmt"
)

var (
	listenPort = flag.String("port", "443", "server listen port")
	syncHost   = flag.String("host", "", "server host")
	syncSer    = flag.Bool("d", false, "server mode")
	syncFold   = flag.String("dir", "./gosync/", "server or client sync fold")
)

func main() {
	flag.Parse()
	if *syncSer {
		addr := fmt.Sprintf(":%s", *listenPort)
		s := NewServer(addr, *syncFold)
		s.Serve()
	} else {
		destination := fmt.Sprintf("%s:%s", *syncHost, *listenPort)
		c := NewClient(destination, *syncFold)
		c.WatchAndSync()
	}
}
