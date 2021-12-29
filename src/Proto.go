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

var fullStatusCode = map[int]string{
	200: "200 OK",
	400: "400 Bad Request",
	404: "404 Not Found",
	408: "408 Request Timeout",
}

type HTTPInterface interface {
	HandleFile(route string, path string)
	HandleFunc(route string, hfunc func(req *http.Request) []byte)
}

type ProtoHTTP struct {
	RouteMap     map[string]respondMethod
	timeOut      int
	requestsLeft int
	interComdata string
}

//Sturct to carry specific respond method for a request --
// it removes unnecessary code for some simple responses(like responding with a html file)
type respondMethod struct {
	RespMethodID string
	RespMethod   interface{}
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
//Respond with custom code -- intended for comples operation, more generally used for POST requests
func (s *ProtoHTTP) HandleFunc(route string, hfunc func(req *http.Request) []byte) {
	s.RouteMap[route] = respondMethod{RespMethodID: "general", RespMethod: hfunc}
}

type StatDatInq interface {
	GetRPMethod(QS string) (respondMethod, error)
	GetKAStats() []interface{}
	ICD(data string)
}

func (s ProtoHTTP) GetKAStats() []interface{} {
	return []interface{}{s.timeOut, s.requestsLeft}
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

func (s *ProtoHTTP) ICD(data string) {
	s.interComdata = data
}

func (s *ProtoHTTP) MakeResponse(request string, rpw *bufio.Writer) string {
	s.interComdata = ""
	rp := ProtoHTTPProcessing{RaqReq: request, ResponseWriter: rpw, DI: s, ParsedReq: &http.Request{URL: &url.URL{}}}
	rp.ReadReqString()
	rp.GetResponseBody()
	rp.MakeHeader()
	rp.WriteResponse()
	return s.interComdata
}

type ProtoHTTPProcessing struct {
	ParsedReq      *http.Request // impl parsed request
	RaqReq         string        //impl raw request *
	Status         string        //impl status string for response
	ResponseWriter *bufio.Writer //impl method to write response back *
	ResponseBody   []byte        //impl Body data being responded
	DI             StatDatInq    //impl Data Inquiry *
	Header         responseHeader
}

func (s *ProtoHTTPProcessing) ReadReqString() {
	req, err := http.ReadRequest(bufio.NewReader((bytes.NewBufferString(s.RaqReq))))
	if err != nil {
		log.Println("ERR parsing request >", err)
		s.Status = fullStatusCode[400]
		s.ParsedReq.URL.Path = "BR"
	} else {
		s.ParsedReq = req
	}
}

func (s *ProtoHTTPProcessing) GetResponseBody() {
	rURL := s.ParsedReq.URL.Path
	v, err := s.DI.GetRPMethod(rURL)
	if rURL == "BR" { //error 400
		s.ResponseBody = v.RespMethod.(func() []byte)()
	} else if err != nil { //error 404
		s.Status = fullStatusCode[404]
		s.ResponseBody = v.RespMethod.(func() []byte)()
	} else { //200 ok
		switch v.RespMethodID {
		case "file":
			s.ResponseBody = v.RespMethod.(func() []byte)()
		case "general":
			s.ResponseBody = v.RespMethod.(func(req *http.Request) []byte)(s.ParsedReq)
		}
	}
}

func (s *ProtoHTTPProcessing) MakeHeader() {
	location, _ := time.LoadLocation("GMT")
	s.Header.add("Date", time.Now().In(location).Format(time.RFC1123))
	s.Header.add("Server", "Spider Server (alpha.ed.1)")
	s.Header.add("Content-Length", strconv.FormatInt(int64(len(s.ResponseBody)), 10))
	CStatus := s.ParsedReq.Header.Get("Connection")
	s.Header.add("Connection", CStatus)
	if CStatus == "keep-alive" {
		s.Header.add("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d", s.DI.GetKAStats()...))
	} else if CStatus == "close" {
		s.DI.ICD("CClose")
	}
}

func (s *ProtoHTTPProcessing) WriteResponse() {
	s.ResponseWriter.WriteString(constructResponse(s.Status, s.Header.headerString, string(s.ResponseBody)))
	s.ResponseWriter.Flush()
}

func constructResponse(status string, header string, body string) string {
	return "HTTP/1.1 " + status + "\r\n" + header + "\r\n" + body
}
