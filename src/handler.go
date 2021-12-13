package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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
	UpdateStatus(statusCode int) string
}

func (s Handler) handle(connection net.Conn) {
	handlerID := getGID()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("#SYS Connection:", connection.RemoteAddr().String(), "GoRoutine:", handlerID)
	rw := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))

	//keep-alive trackers
	startTime := time.Now()
	requestsLeft := 6
	timeOut := 5

	//request inwards
	for {
		req, Reqerr := http.ReadRequest(rw.Reader)

		elap := time.Since(startTime).Seconds()
		if elap >= float64(timeOut) {
			break
		}

		if Reqerr != nil {
			log.Println("Error!", "GoRoutine:  -", handlerID, "-", Reqerr)
			return
		}
		fmt.Println(req.Proto, req.Method, req.RequestURI)
		for k, v := range req.Header {
			fmt.Println("	", k, v)
		}

		RStatusCode := 200
		RStatus := fullStatusCode[RStatusCode]

		//body retrive
		respBody := []byte{}
		if k, v := s.RouteMap[req.RequestURI]; v {
			respBody = k(*req)
		} else {
			respBody = s.RouteMap["CNF"](*req)
			RStatusCode, RStatus = s.UpdateStatus(404)
		}

		header := make(map[string][]string)

		//h - connection
		ConnHeader := req.Header.Get("Connection")
		if ConnHeader == "keep-alive" {
			header["Connection"] = []string{"keep-alive"}
			header["Keep-Alive"] = []string{fmt.Sprintf("timeout=%d, max=%d", timeOut, requestsLeft)}
		}

		Resp := &http.Response{
			Status:        RStatus,
			StatusCode:    RStatusCode,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          io.NopCloser(bytes.NewBuffer(respBody)),
			ContentLength: int64(len(respBody)),
			Header:        header,
		}
		req.Response = Resp
		req.Response.Write(rw.Writer)
		rw.Flush()

		requestsLeft--

		if ConnHeader == "close" || requestsLeft == 0 {
			break
		}
	}
	connection.Close()
	fmt.Println("#SYS Complete!", "GoRoutine:", handlerID, "→ Closed")
	fmt.Println(strings.Repeat("-", 50))
}

func (s *Handler) HandleFunc(route string, hfunc func(req http.Request) []byte) {
	s.RouteMap[route] = hfunc
}

func (Handler) UpdateStatus(statusCode int) (int, string) {
	return statusCode, fullStatusCode[statusCode]
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
