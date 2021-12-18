package main

import "time"

type responseHeader struct {
	headerSetters

	headerString string
}

type headerSetters interface {
	add(header string, value string)
}

func (s *responseHeader) add(header string, value string) {
	s.headerString += header + ": " + value + "\r\n"
}

func NewHeader() *responseHeader {
	h := new(responseHeader)
	location, _ := time.LoadLocation("GMT")
	h.add("Date", time.Now().In(location).Format(time.RFC1123))
	h.add("Server", "Spider Server (dev.ed.V15)")
	return h
}
