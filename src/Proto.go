package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

//HTTP Section
var fullStatusCode = map[int]string{
	200: "200 OK",
	400: "400 Bad Request",
	404: "404 Not Found",
	408: "408 Request Timeout",
}

//HTTP Protocol
type ProtoHTTP struct {
	RouteMap   map[string]respondMethod
	LTData     LifeTimeData
	returnData string
}

//Sturct to carry specific respond method for a request --
// it removes unnecessary code for some simple responses(like responding with a html file)
type respondMethod struct {
	RespMethodID string
	RespMethod   interface{}
}

//Selective functions that are exposed as methods of the server to provide customizability
type HTTPInterface interface {
	HandleFile(route string, path string)
	HandleFunc(route string, hfunc func(req *http.Request) []byte)
}

//HandleFile
//Respond with data from a file
func (s *ProtoHTTP) HandleFile(route string, path string) {
	s.RouteMap[route] = respondMethod{RespMethodID: "file", RespMethod: func() []byte {
		data, _ := os.ReadFile(path)
		return data
	}}
}

//HandleFunc
//Respond with custom code -- intended for complex operation, more generally used for POST requests
func (s *ProtoHTTP) HandleFunc(route string, hfunc func(req *http.Request) []byte) {
	s.RouteMap[route] = respondMethod{RespMethodID: "general", RespMethod: hfunc}
}

//Methods used by the ProtoHTTPProcessing (true handler) to communicate with the Main Handler and the HTTP Protocol struct
type HigherLevelDataQuery interface {
	GetRPMethod(QS string) (respondMethod, error)
	ReturnData(data string)
}

func (s *ProtoHTTP) GetRPMethod(QS string) (respondMethod, error) {
	var rto respondMethod
	var ErrorString error
	if v, exist := s.RouteMap[QS]; exist {
		rto = v
	} else {
		log.Println("E404 Not Found")
		rto = s.RouteMap["CNF"]
		ErrorString = CreateError("404 Not Found")
	}
	return rto, ErrorString
}

func (s *ProtoHTTP) ReturnData(data string) {
	s.returnData = data
}

func (s *ProtoHTTP) MakeResponse(request string, rpw *bufio.Writer) string {
	s.returnData = ""
	tOut, Max := s.LTData.GetLifeTime()
	rp := HTTPRespHandler{RawReq: request, ParsedReq: &http.Request{URL: &url.URL{}}, DQ: s, TimeOut: tOut, MaxRequests: Max}
	rp.Header.hMap = make(map[string]string)
	rp.ReadReqString()
	rp.GetResponseBody()
	rp.MakeHeader()
	rpw.WriteString(constructResponse(rp.Status, rp.Header.String(), string(rp.ResponseBody)))
	rpw.Flush()
	return s.returnData
}

type HTTPRespHandler struct {
	ParsedReq    *http.Request        // impl parsed request
	RawReq       string               //impl raw request *
	Status       string               //impl status string for response
	ResponseBody []byte               //impl Body data being responded
	DQ           HigherLevelDataQuery //impl Data Inquiry *
	Header       responseHeader
	TimeOut      int
	MaxRequests  int
}

func (s *HTTPRespHandler) ReadReqString() error {
	req, err := http.ReadRequest(bufio.NewReader((bytes.NewBufferString(s.RawReq))))
	if err != nil {
		log.Println("ERR parsing request >", err)
		s.Status = fullStatusCode[400]
		s.ParsedReq.URL.Path = "BR"
		return CreateError("Cannot parse request")
	}
	s.ParsedReq = req
	return nil
}

func (s *HTTPRespHandler) GetResponseBody() {
	rURL := s.ParsedReq.URL.Path
	v, err := s.DQ.GetRPMethod(rURL)
	if rURL == "BR" { //error 400
		s.ResponseBody = v.RespMethod.(func() []byte)()
	} else if err != nil { //error 404
		s.Status = fullStatusCode[404]
		s.ResponseBody = v.RespMethod.(func() []byte)()
	} else { //200 ok
		s.Status = fullStatusCode[200]
		switch v.RespMethodID {
		case "file":
			s.ResponseBody = v.RespMethod.(func() []byte)()
		case "general":
			s.ResponseBody = v.RespMethod.(func(req *http.Request) []byte)(s.ParsedReq)
		}
	}
}

func (s *HTTPRespHandler) MakeHeader() {
	location, _ := time.LoadLocation("GMT")
	s.Header.add("Date", time.Now().In(location).Format(time.RFC1123))
	s.Header.add("Server", "Spider Server (alpha.ed.2.3)")
	s.Header.add("Content-Length", strconv.FormatInt(int64(len(s.ResponseBody)), 10))
	CStatus := s.ParsedReq.Header.Get("Connection")
	s.Header.add("Connection", CStatus)
	if CStatus == "keep-alive" {
		s.Header.add("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d", s.TimeOut, s.MaxRequests))
	} else if CStatus == "close" {
		s.DQ.ReturnData("CClose")
	}
}

func constructResponse(status string, header string, body string) string {
	return "HTTP/1.1 " + status + "\r\n" + header + "\r\n" + body
}
