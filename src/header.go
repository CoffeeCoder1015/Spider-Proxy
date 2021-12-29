package main

type responseHeader struct {
	headerString string
}

func (s *responseHeader) add(header string, value string) {
	s.headerString += header + ": " + value + "\r\n"
}
