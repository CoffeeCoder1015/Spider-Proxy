package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

/*
Init stage obj var
calls on init var
*/

//TCP Server
func TCPServer(addr string, port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Println("server starting err", err)
		return false
	}
	HStart := InitHandling()
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
