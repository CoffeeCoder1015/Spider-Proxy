package main

import (
	"bytes"
	"fmt"
	"io"
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
	s := NewHTTPProxyServer()
	s.HTTPProxy.HandleFile("/", "content/index.html")
	s.HTTPProxy.HandleFile("/stuff.js", "content/stuff.js")
	s.HTTPProxy.HandleFile("/CloseButton.png", "content/CloseButton.png")

	ansChan := make(chan string)

	s.HTTPProxy.HandleFunc("/math", func(req *http.Request) []byte {
		Q := req.URL.Query().Get("eq")
		ans := doMath(Q)
		fmt.Println(Q, ans)
		go func() { ansChan <- ans }()
		fmt.Println(ansChan)
		data, _ := os.ReadFile("content/app/QueryStringTester.html")
		return data
	})
	s.HTTPProxy.HandleFunc("/math.ans", func(req *http.Request) []byte {
		ans := <-ansChan
		fmt.Println(ans)
		return []byte(ans)
	})
	s.HTTPProxy.HandleFile("/math/getAns.js", "content/app/getAns.js")
	s.HTTPProxy.RedirectFunction("http://gib/", func(req *http.Request) []byte {
		c, _ := http.Get("http://localhost:7123")
		data, _ := io.ReadAll(c.Body)
		return data
	})
	s.HTTPProxy.ResponseHeaderOveride(HeaderManip{Field: "Server", Value: "Spider Proxy alph v1 (Powered by Spider Server alpha v2)"})
	s.TCPServer("", 8080)
}
