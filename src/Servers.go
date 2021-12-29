package main

import (
	"fmt"
	"log"
	"net"
)

type SpiderServer struct {
	Handler
	HTTP HTTPInterface
}

type TCPServer interface {
	TCPServer(addr string, port int) bool
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

func NewHTTPServer() *SpiderServer {
	Proto := ProtoHTTP{timeOut: 5, requestsLeft: 10, RouteMap: make(map[string]respondMethod)}
	Proto.HandleFile("CNF", "content/Errors/CNF.html")
	Proto.HandleFile("BR", "content/Errors/BR.html")
	Server := new(SpiderServer)
	Server.HTTP = &Proto
	Server.HandlingInterface = &Proto
	return Server
}
