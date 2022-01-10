package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	HandlingInterface HandleInterface
	tlsConfig         tls.Config
	timeOut           int
	requestsLeft      int
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
	return s.timeOut, s.requestsLeft
}

func (s Handler) handle(connection net.Conn) {
	handlerID := getGID()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("#SYS Connection:", connection.RemoteAddr().String(), "GoRoutine:", handlerID, time.Now())
	rw := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))

	TransferBuf := new(bytes.Buffer)
	tlsConn := tls.Server(tempConn{Conn: connection, reader: io.TeeReader(connection, TransferBuf)}, &s.tlsConfig)
	err := tlsConn.Handshake()

	ConnOpen := true

	if err != nil {
		fmt.Println("Err on initalising TLS >", err)
		Data, breadFullErr := BufioReadFull(*bufio.NewReader(connection))
		if breadFullErr != nil {
			log.Println("FR Err >", breadFullErr)
		}
		TransferBuf.Write(Data)
		intCom := s.HandlingInterface.MakeResponse(TransferBuf.String(), rw.Writer)
		s.requestsLeft--
		if intCom == "CClose" {
			ConnOpen = false
		}
	} else {
		connection = tlsConn
		rw = bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))
	}

	//keep-alive trackers
	timeOut := s.timeOut
	tOutInDura := time.Duration(timeOut) * time.Second
	reqChan := make(chan string)
	for ConnOpen {
		go func() {
			DataBuf, Reqerr := BufioReadFull(*rw.Reader)
			if !ConnOpen {
				return
			}
			if Reqerr != nil {
				log.Println("Error! GoRoutine:  -", handlerID, " > Request Error > ", Reqerr)
			} else {
				reqChan <- string(DataBuf)
			}
		}()
		select {
		case req := <-reqChan:
			fmt.Println(req)
			fmt.Println(strings.Repeat("-", 10))
			intCom := s.HandlingInterface.MakeResponse(req, rw.Writer)
			s.requestsLeft--
			if intCom == "CClose" || s.requestsLeft == 0 {
				ConnOpen = false
			}
		case time := <-time.After(tOutInDura):
			fmt.Println(time, "TIMEOUT", handlerID)
			ConnOpen = false
		}

	}

	defer connection.Close()
	fmt.Println("#SYS Complete!", "GoRoutine:", handlerID, "→ Closed")
	fmt.Println(strings.Repeat("-", 50))
}

func BufioReadFull(r bufio.Reader) ([]byte, error) {
	DataBuf := make([]byte, r.Size())
	_, Reqerr := r.Read(DataBuf)
	if dInBuf := r.Buffered(); dInBuf > 0 {
		dataInBuffer, pkErr := r.Peek(r.Buffered())
		if pkErr != nil {
			log.Println("Peek Error > ", pkErr)
		}
		DataBuf = append(DataBuf, dataInBuffer...)
		r.Discard(r.Buffered())
	}
	return DataBuf, Reqerr
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

type tempConn struct {
	net.Conn
	reader io.Reader
}

func (conn tempConn) Read(p []byte) (int, error) { return conn.reader.Read(p) }

//debug
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
