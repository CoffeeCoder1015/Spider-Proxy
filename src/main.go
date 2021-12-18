package main

//TODO -  RENAME PACKAGE

func main() {
	s := NewServer()
	s.HandleFile("/", "content/index.html")
	s.HandleFile("/stuff.js", "content/stuff.js")
	s.HandleFile("/CloseButton.png", "content/CloseButton.png")
	s.TCPServer("", 8080)
}
