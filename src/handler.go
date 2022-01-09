package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	HandlingInterface HandleInterface
	timeOut           int
	requestsPerHandle int
}

//indicates accpetted function
//the standard route by which a protocol creates a response for the handler to call -- it is not how something interacts with the handler, it is how the handler interacts with something else
type HandleInterface interface {
	MakeResponse(request string, rpw *bufio.Writer) string
}

//exposed function
// this is how the handler interacts with something else
type LifeTimeData interface {
	GetLifeTime() (int, int)
}

func (s Handler) GetLifeTime() (int, int) {
	return s.timeOut, s.requestsPerHandle
}

func (s Handler) handle(connection net.Conn) {
	handlerID := getGID()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("#SYS Connection:", connection.RemoteAddr().String(), "GoRoutine:", handlerID, time.Now())
	rw := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))

	//keep-alive trackers
	requestsLeft := s.requestsPerHandle
	timeOut := s.timeOut
	tOutInDura := time.Duration(timeOut) * time.Second
	reqChan := make(chan string)
	ConnectionClosed := true

	for ConnectionClosed {
		go func() {
			DataBuf := make([]byte, rw.Reader.Size())
			_, Reqerr := rw.Read(DataBuf)
			if dInBuf := rw.Reader.Buffered(); dInBuf > 0 {
				dataInBuffer, pkErr := rw.Peek(rw.Reader.Buffered())
				if pkErr != nil {
					log.Println("Peek Error > ", pkErr)
				}
				DataBuf = append(DataBuf, dataInBuffer...)
				rw.Discard(rw.Reader.Buffered())
			}
			if !ConnectionClosed {
				return
			}
			if Reqerr != nil {
				log.Println("Error! GoRoutine:  -", handlerID, " > Request Error > ", Reqerr)
				if Reqerr.Error() == "tls: first record does not look like a TLS handshake" {
					fmt.Println("switch to no tls")
					return
				}
			} else {
				reqChan <- string(DataBuf)
			}
		}()
		select {
		case req := <-reqChan:
			fmt.Println(req)
			fmt.Println(strings.Repeat("-", 10))
			var intCom string
			intCom = s.HandlingInterface.MakeResponse(req, rw.Writer)
			requestsLeft--
			if intCom == "CClose" || requestsLeft == 0 {
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
func (s *CustomError) Error() string {
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
