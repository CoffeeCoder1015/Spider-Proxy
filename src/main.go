package main

import (
	"net/http"
	"os"
)

//TODO -  RENAME PACKAGE

func main() {
	s := NewServer()
	s.HandleFunc("/", func(req http.Request) []byte {
		file, _ := os.ReadFile("content/index.html")
		return file
	})
	s.HandleFunc("/stuff.js", func(req http.Request) []byte {
		file, _ := os.ReadFile("content/stuff.js")
		return file
	})
	s.HandleFunc("/CloseButton.png", func(req http.Request) []byte {
		file, _ := os.ReadFile("content/CloseButton.png")
		return file
	})
	s.TCPServer("", 8080)
}
