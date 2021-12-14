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
	"time"
)

var fullStatusCode = map[int]string{
	200: "200 OK",
	404: "404 Not Found",
}

type Handler struct {
	HandleOperators

	RouteMap map[string]func(http.Request) []byte
}

type HandleOperators interface {
	handle(connection net.Conn)
	HandleFunc(route string, hfunc func(req http.Request) []byte)
}

func (s Handler) handle(connection net.Conn) {
	handlerID := getGID()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("#SYS Connection:", connection.RemoteAddr().String(), "GoRoutine:", handlerID)
	rw := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))

	//keep-alive trackers
	requestsLeft := 100
	timeOut := 5
	tOutInDura := time.Duration(timeOut) * time.Second
	reqChan := make(chan *http.Request)
KeepAliveLoop:
	for {

		go func() {
			req, Reqerr := http.ReadRequest(rw.Reader)
			if Reqerr != nil {
				log.Println("Error!", "GoRoutine:  -", handlerID, "-", Reqerr)
			}
			reqChan <- req
		}()

		select {
		case req := <-reqChan:
			//data disp
			fmt.Println(req.Proto, req.Method, req.RequestURI)
			for k, v := range req.Header {
				fmt.Println("	", k, v)
			}
			fmt.Println(time.Now())
			fmt.Println("	", strings.Repeat("-", 20))

			RStatus := fullStatusCode[200]

			//body retrive
			respBody := []byte{}
			if k, v := s.RouteMap[req.RequestURI]; v {
				respBody = k(*req)
			} else {
				respBody = s.RouteMap["CNF"](*req)
				RStatus = fullStatusCode[404]
			}

			header := make(map[string]string)
			//h - date
			location, _ := time.LoadLocation("GMT")
			header["Date"] = time.Now().In(location).String()
			//h - server prod
			header["Server"] = "Spider Server (bV12)"
			//h - cont length
			header["Content-Length"] = strconv.FormatInt(int64(len(respBody)), 10)

			//h - connection
			ConnHeader := req.Header.Get("Connection")
			if ConnHeader == "keep-alive" {
				header["Connection"] = "keep-alive"
				header["Keep-Alive"] = fmt.Sprintf("timeout=%d, max=%d", timeOut, requestsLeft)
			}

			RespBlk := []string{"HTTP/1.1 " + RStatus}
			for k, v := range header {
				RespBlk = append(RespBlk, fmt.Sprintf("%s: %s", k, v))
			}
			Resp := strings.Join(RespBlk, "\r\n") + "\r\n\r\n" + string(respBody)
			rw.WriteString(Resp)
			rw.Flush()

			requestsLeft--
			if ConnHeader == "close" || requestsLeft == 0 {
				break KeepAliveLoop
			}
		case time := <-time.After(tOutInDura):
			fmt.Println(time, "TIMEOUT")
			break KeepAliveLoop
		}

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
