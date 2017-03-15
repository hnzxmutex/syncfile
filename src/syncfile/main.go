package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

var (
	listenPort = flag.String("port", "443", "server listen port")
	syncHost   = flag.String("host", "", "server host")
	syncSer    = flag.Bool("d", false, "server mode")
	syncFold   = flag.String("dir", "./gosync/", "server or client sync fold")
	ignoreFile = flag.String("i", "./ignore.ini", "ignore file")
	password   = flag.String("p", "tgideas", "password")
)

func main() {
	flag.Parse()
	//ignore list

	ignoreData, err := ioutil.ReadFile(*ignoreFile)
	var ignoreListString []string
	if err != nil {
		log.Println("Warning:read ignore file error,sync all file")
	} else {
		ignoreString := strings.TrimLeft(string(ignoreData), " ")
		ignoreString = strings.Replace(ignoreString, "\r\n", "\n", -1)
		ignoreString = strings.Replace(ignoreString, "\r", "\n", -1)
		ignoreString = strings.Trim(string(ignoreData), "\n")
		ignoreListString = strings.Split(ignoreString, "\n")
		if len(ignoreListString) == 1 && ignoreListString[0] == "" {
			ignoreListString = []string{}
		}
	}

	ignoreList := checkIgnoreList(ignoreListString)
	if *syncSer {
		addr := fmt.Sprintf(":%s", *listenPort)
		s := NewServer(addr, *syncFold, *password, ignoreList)
		s.Serve()
	} else {
		destination := fmt.Sprintf("%s:%s", *syncHost, *listenPort)
		c := NewClient(destination, *syncFold, *password, ignoreList)
		c.WatchAndSync()
	}
}

func checkIgnoreList(ignoreList []string) []*regexp.Regexp {
	regList := make([]*regexp.Regexp, 0, len(ignoreList))
	for _, p := range ignoreList {
		p = strings.TrimLeft(p, " ")
		if ([]rune(p))[0] == '#' {
			continue
		}
		regList = append(regList, regexp.MustCompile(p))
	}
	return regList
}
