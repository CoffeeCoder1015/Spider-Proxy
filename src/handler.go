package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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
	ProxyOut

	RouteMap  map[string]respondMethod
	ProxyMode bool
}

type HandleOperators interface {
	handle(connection net.Conn)
	GetResponseBody(req http.Request) []byte
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
	ConnectionClosed := true
	for ConnectionClosed {
		go func() {
			req, Reqerr := http.ReadRequest(rw.Reader)
			if !ConnectionClosed {
				return
			}
			if Reqerr != nil {
				log.Println("Error!", "GoRoutine:  -", handlerID, "-", Reqerr)
			} else {
				reqChan <- req
			}
		}()
		select {
		case req := <-reqChan:
			//data disp
			fmt.Println(req.Proto, req.Method, req.URL, req.URL.Query())
			fmt.Println(req.RequestURI, req.URL.Path)
			for k, v := range req.Header {
				fmt.Println("	", k, v)
			}
			fmt.Println(time.Now())
			fmt.Println("	", strings.Repeat("-", 20))

			RStatus := fullStatusCode[200]

			//body retrive
			respBody, RespErr := s.GetResponseBody(req)
			if RespErr.Error() == "404 Not Found" {
				log.Println(RespErr, req.RequestURI, handlerID)
				RStatus = fullStatusCode[404]
			} else {
				log.Println(req.RequestURI, RStatus)
			}

			header := NewHeader()
			header.add("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))

			ConnHeader := req.Header.Get("Connection")
			if ConnHeader == "keep-alive" {
				header.add("Connection", "keep-alive")
				header.add("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d", timeOut, requestsLeft))
			}
			Resp := "HTTP/1.1 " + RStatus + "\r\n" + header.headerString + "\r\n" + string(respBody)
			rw.WriteString(Resp)
			rw.Flush()

			requestsLeft--
			if ConnHeader == "close" || requestsLeft == 0 {
				ConnectionClosed = false
			}
		case time := <-time.After(tOutInDura):
			fmt.Println(time, "TIMEOUT", handlerID)
			ConnectionClosed = false
		}

	}

	defer connection.Close()
	fmt.Println("#SYS Complete!", "GoRoutine:", handlerID, "→ Closed")
	fmt.Println(strings.Repeat("-", 50))
}

func (s *Handler) GetResponseBody(req *http.Request) ([]byte, error) {
	respBody := []byte{}
	ErrorString := &CustomError{}
	proxySuccess := false
	if s.ProxyMode {
		r, err := s.ProxyResponse(req)
		if err == nil {
			respBody = r
			proxySuccess = true
		} else {
			log.Println("PROXY", err)
		}
	}
	if !proxySuccess {
		if v, exist := s.RouteMap[req.URL.Path]; exist {
			switch v.RespMethodID {
			case "file":
				respBody = v.RespMethod.(func() []byte)()
			case "general":
				respBody = v.RespMethod.(func(req *http.Request) []byte)(req)
			}
		} else {
			fmt.Println("Proc")
			respBody = s.RouteMap["CNF"].RespMethod.(func() []byte)()
			ErrorString = CreateError("404 Not Found")
		}
	}

	return respBody, ErrorString
}

func (s *Handler) ProxyResponse(req *http.Request) ([]byte, error) {
	if !isUrl(req.RequestURI) {
		return []byte{}, CreateError("Not a URL")
	}
	return s.ProxReq(req), nil
}

//Sturct to carry specific respond method for a request --
// it removes unnecessary code for some simple responses(like responding with a html file)
type respondMethod struct {
	RespMethodID string
	RespMethod   interface{}
}

//HandleFile
//Respond with data from a file
func (s *Handler) HandleFile(route string, path string) {
	s.RouteMap[route] = respondMethod{RespMethodID: "file", RespMethod: func() []byte {
		data, _ := os.ReadFile(path)
		return data
	}}
}

//HandleFunc
//Respond with custom code -- intended for comples operation, more generally used for POST requests
func (s *Handler) HandleFunc(route string, hfunc func(req *http.Request) []byte) {
	s.RouteMap[route] = respondMethod{RespMethodID: "general", RespMethod: hfunc}
}

//Renamed code from erros.New to generate errors
//To get a Error value
func CreateError(text string) *CustomError {
	return &CustomError{s: text}
}

//struct that is needed to carry the Error info
type CustomError struct {
	s string
}

// Defined in *builtins* that a error type is interface w/ function: Error() string
func (s CustomError) Error() string {
	return s.s
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
