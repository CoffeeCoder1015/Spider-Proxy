package main

import (
	"fmt"
	"log"
	"net"
)

type ServerSkeleton struct {
	Handler
}

//TCP Server
func (s *ServerSkeleton) TCPServer(addr string, port int) bool {
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

type HTTPServer struct {
	ServerSkeleton
	HTTP HTTPInterface
}

func NewHTTPServer() *HTTPServer {
	Server := new(HTTPServer)
	Server.timeOut = 5
	Server.requestsPerHandle = 10

	Proto := ProtoHTTP{LTData: Server, RouteMap: make(map[string]respondMethod)}
	Proto.HandleFile("CNF", "content/Errors/CNF.html")
	Proto.HandleFile("BR", "content/Errors/BR.html")
	Server.HTTP = &Proto
	Server.HandlingInterface = &Proto

	return Server
}

type HTTPProxyServer struct {
	ServerSkeleton
	HTTPProxy HTTPProxyInterface
}

func NewHTTPProxyServer() *HTTPProxyServer {
	Server := new(HTTPProxyServer)
	Server.timeOut = 5
	Server.requestsPerHandle = 10

	Proto := ProtoHTTProxy{ProtoHTTP: ProtoHTTP{LTData: Server, RouteMap: make(map[string]respondMethod)}, URLOverideMap: make(map[string]respondMethod)}
	Proto.HandleFile("CNF", "content/Errors/CNF.html")
	Proto.HandleFile("BR", "content/Errors/BR.html")
	Server.HTTPProxy = &Proto
	Server.HandlingInterface = &Proto
	return Server
}
