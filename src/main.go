package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

//TODO -  RENAME PACKAGE

func doMath(equation string) string {
	cmd := exec.Command("content/app/app.exe")
	var out bytes.Buffer
	cmd.Stdin = bytes.NewBufferString(equation)
	cmd.Stdout = &out
	cmd.Run()
	return out.String()
}

func main() {
	s := NewServer(true)
	s.HandleFile("/", "content/index.html")
	s.HandleFile("/stuff.js", "content/stuff.js")
	s.HandleFile("/CloseButton.png", "content/CloseButton.png")

	ansChan := make(chan string)

	s.HandleFunc("/math", func(req *http.Request) []byte {
		Q := req.URL.Query().Get("eq")
		ans := doMath(Q)
		fmt.Println(Q, ans)
		go func() { ansChan <- ans }()
		fmt.Println(ansChan)
		data, _ := os.ReadFile("content/app/QueryStringTester.html")
		return data
	})
	s.HandleFunc("/math.ans", func(req *http.Request) []byte {
		ans := <-ansChan
		fmt.Println(ans)
		return []byte(ans)
	})
	s.HandleFile("/math/getAns.js", "content/app/getAns.js")
	s.TCPServer("", 8080)
}
