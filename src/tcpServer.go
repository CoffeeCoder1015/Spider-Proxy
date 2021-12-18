package main

import (
	"fmt"
	"log"
	"net"
)

type ServerCore interface {
	TCPServer(addr string, port int) bool
}

type SpiderServer struct {
	ServerCore
	Handler

	ProxyMode bool
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
	Serv.RouteMap = make(map[string]respondMethod)
	Serv.HandleFile("CNF", "content/CNF.html")
	return Serv
}
