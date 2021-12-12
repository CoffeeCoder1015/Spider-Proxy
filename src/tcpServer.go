package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type ServerCore interface {
	TCPServer(addr string, port int) bool
}

type SpiderServer struct {
	ServerCore
	Handler
}

//TCP Server
func (s *SpiderServer) TCPServer(addr string, port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Println("server starting err", err)
		return false
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			return false
		}
		go s.handle(conn)
	}
}

func NewServer() *SpiderServer {
	Serv := new(SpiderServer)
	Serv.RouteMap = make(map[string]interface{})
	Serv.RouteMap["CNF"] = func(req http.Request) []byte {
		contentNotFoundFile, cnffReaderr := os.ReadFile("CNF.html")
		if cnffReaderr != nil {
			log.Panicln("ReadErr:", cnffReaderr)
		}
		return contentNotFoundFile
	}
	return Serv
}
