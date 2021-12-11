package main

import (
	"net/http"
	"os"
)

//TODO -  RENAME PACKAGE

func main() {
	s := SpiderServer{}
	s.HandleFunc("/", func(req http.Request) []byte {
		file, _ := os.ReadFile("content/index.html")
		return file
	})
	s.TCPServer("", 8080)
}
