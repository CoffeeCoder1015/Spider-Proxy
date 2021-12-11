package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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

func (s Handler) handle(handlerID string, connection net.Conn) {
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

	//Connection - handling
	if req.Header.Get("Connection") == "keep-alive" {
		for {
			kaReq, krqErr := http.ReadRequest(rw.Reader)
			if krqErr != nil {
				log.Println("Error!", "GoRoutine:  -", handlerID, "-", krqErr)
			}
			fmt.Println(kaReq.Method, kaReq.Proto)
			fmt.Println("Req url:", kaReq.RequestURI, "Requested content: ")
			for k, v := range kaReq.Header {
				fmt.Println("KA Loop", handlerID, k, v)
			}
		}
	}

	connection.Close()
	fmt.Println("#SYS Complete!", "GoRoutine:", handlerID, "→ Closed")
	fmt.Println(strings.Repeat("-", 50))
}

func (s *Handler) HandleFunc(route string, hfunc func(req http.Request) []byte) {
	s.RouteMap[route] = hfunc
}

func InitHandling() *Handler {
	handleStart := new(Handler)
	handleStart.RouteMap = map[string]interface{}{}
	handleStart.RouteMap["CNF"] = func(req http.Request) []byte {
		contentNotFoundFile, cnffReaderr := os.ReadFile("CNF.html")
		if cnffReaderr != nil {
			log.Panicln("ReadErr:", cnffReaderr)
		}
		return contentNotFoundFile
	}
	return handleStart
}
