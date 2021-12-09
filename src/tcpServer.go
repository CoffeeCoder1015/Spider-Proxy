package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

/*
Init stage obj var
calls on init var
*/

type Handler struct {
	RouteMap map[string]string
}

type Handle interface {
	handle(connection net.Conn)
	GET(req *http.Request, respW *bufio.Writer)
}

func (s Handler) handle(handlerID string, connection net.Conn) {
	fmt.Println("#SYS Connection:", connection.RemoteAddr().String(), "GoRoutine:", handlerID)
	rw := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))
	req, Reqerr := http.ReadRequest(rw.Reader)
	if Reqerr != nil {
		log.Println("Error!", "GoRoutine:  -", handlerID, "-", Reqerr)
		return
	}

	switch req.Method {
	case "GET":
		s.GET(req, rw.Writer)
	}

	connection.Close()
	fmt.Println("#SYS Complete!", "GoRoutine:", handlerID, "→ Closed")
	fmt.Println(strings.Repeat("-", 50))
}

func (s Handler) GET(req *http.Request, respW *bufio.Writer) {
	fmt.Printf("req.Method: %v\n", req.Method)
	fmt.Printf("req.RequestURI: %v\n", req.RequestURI)
	fmt.Printf("req.Proto: %v\n", req.Proto)
	fmt.Println(req.ProtoMajor, req.ProtoMinor)
	for k, v := range req.Header {
		fmt.Println(k, v)
	}
}

func InitHandling(routingMap map[string]string) *Handler {
	handleStart := new(Handler)
	handleStart.RouteMap = routingMap
	handleStart.RouteMap["CNF"] = "CNF.html"
	return handleStart
}

//TCP Server
func TCPServer(addr string, port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Println("server starting err", err)
		return false
	}
	HStart := InitHandling(map[string]string{"/": "index.html"})
	Hid := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			return false
		}
		go HStart.handle(strconv.FormatInt(int64(Hid), 10), conn)
		Hid++
	}
}
