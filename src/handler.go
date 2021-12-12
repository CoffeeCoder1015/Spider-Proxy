package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

type Handler struct {
	HandleOperators

	RouteMap map[string]interface{}
}

type HandleOperators interface {
	handle(connection net.Conn)
	HandleFunc(route string, hfunc func(req http.Request) []byte)
}

func (s Handler) handle(connection net.Conn) {
	handlerID := getGID()
	fmt.Println("#SYS Connection:", connection.RemoteAddr().String(), "GoRoutine:", handlerID)
	rw := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))

	req, Reqerr := http.ReadRequest(rw.Reader)
	if Reqerr != nil {
		log.Println("Error!", "GoRoutine:  -", handlerID, "-", Reqerr)
		return
	}
	fmt.Println(req.Method, req.Proto)
	fmt.Println("Req url:", req.RequestURI, "Requested content: ")
	for k, v := range req.Header {
		fmt.Println(k, v)
	}

	connection.Close()
	fmt.Println("#SYS Complete!", "GoRoutine:", handlerID, "→ Closed")
	fmt.Println(strings.Repeat("-", 50))
}

func (s *Handler) HandleFunc(route string, hfunc func(req http.Request) []byte) {
	s.RouteMap[route] = hfunc
}

//debug
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
