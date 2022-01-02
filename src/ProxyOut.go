package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

//Handling protocol for a HTTP Proxy
type ProtoHTTProxy struct {
	ProtoHTTP
	URLOverideMap map[string]respondMethod
	ReqHeaderOV   []HeaderManip
	RpHeaderOV    []HeaderManip
}

type HeaderManip struct {
	Field string
	Value string
	Del   bool
}

type HTTPProxyInterface interface {
	HTTPInterface
	RequestHeaderOveride(header HeaderManip)
	ResponseHeaderOveride(header HeaderManip)
	RedirectFile(route string, path string)
	RedirectFunction(route string, hfunc func(req *http.Request) []byte)
}

func (s *ProtoHTTProxy) RedirectFile(route string, path string) {
	s.URLOverideMap[route] = respondMethod{RespMethodID: "file", RespMethod: func() []byte {
		data, _ := os.ReadFile(path)
		return data
	}}
}

func (s *ProtoHTTProxy) RedirectFunction(route string, hfunc func(req *http.Request) []byte) {
	s.URLOverideMap[route] = respondMethod{RespMethodID: "general", RespMethod: hfunc}
}

func (s *ProtoHTTProxy) RequestHeaderOveride(header HeaderManip) {
	s.ReqHeaderOV = append(s.ReqHeaderOV, header)
}
func (s *ProtoHTTProxy) ResponseHeaderOveride(header HeaderManip) {
	s.RpHeaderOV = append(s.RpHeaderOV, header)
}

type ProxyHLDQ interface {
	GetOVHeaders(queryFor string) []HeaderManip
	GetRedirects(QS string) (respondMethod, error)
}

func (s ProtoHTTProxy) GetOVHeaders(queryFor string) []HeaderManip {
	switch queryFor {
	case "req":
		return s.ReqHeaderOV
	case "res":
		return s.RpHeaderOV
	}
	return nil
}

func (s *ProtoHTTProxy) GetRedirects(QS string) (respondMethod, error) {
	var rto respondMethod
	var ErrorString error
	v, exist := s.URLOverideMap[QS]
	if exist {
		rto = v
	} else {
		ErrorString = CreateError("No Overide")
	}
	fmt.Println(QS, exist)
	return rto, ErrorString
}

func (s *ProtoHTTProxy) MakeResponse(request string, rpw *bufio.Writer) string {
	s.returnData = ""
	tOut, Max := s.LTData.GetLifeTime()
	ParsedRequestInit := &http.Request{URL: &url.URL{}}
	rp := HTTPProxyRespHandler{HTTPRespHandler: HTTPRespHandler{RawReq: request, ParsedReq: ParsedRequestInit, DQ: s, TimeOut: tOut, MaxRequests: Max}, ProxyHLDQ: s}
	rp.Header.hMap = make(map[string]string)
	err := rp.ReadReqString()
	if err == nil {
		rp.isRequestToProxy()
	} else {
		rp.isLocal = true
	}
	rp.ProxyRequest()
	rp.MakeHeader()
	if !rp.isLocal {
		rp.RespHeadOveride()
	}
	rpw.WriteString(constructResponse(rp.Status, rp.Header.String(), string(rp.ResponseBody)))
	rpw.Flush()
	return s.returnData
}

type HTTPProxyRespHandler struct {
	client  http.Client
	isLocal bool
	HTTPRespHandler
	ProxyHLDQ ProxyHLDQ
}

func (s *HTTPProxyRespHandler) isRequestToProxy() {
	if string(s.ParsedReq.RequestURI[0]) == "/" {
		s.isLocal = true
	} else {
		s.isLocal = false
	}
}

func (s *HTTPProxyRespHandler) ProxyRequest() {
	fmt.Println(s.ParsedReq.URL.Scheme+"://"+s.ParsedReq.URL.Host+s.ParsedReq.URL.Path, s.ParsedReq.RequestURI)
	rpm, e := s.ProxyHLDQ.GetRedirects(s.ParsedReq.URL.Scheme + "://" + s.ParsedReq.URL.Host + s.ParsedReq.URL.Path)
	switch rpm.RespMethodID {
	case "general":
		s.ResponseBody = rpm.RespMethod.(func(req *http.Request) []byte)(s.ParsedReq)
	case "file":
		s.ResponseBody = rpm.RespMethod.(func() []byte)()
	}
	if e == nil {
		return
	}
	if s.isLocal {
		s.GetResponseBody()
	} else {
		s.ParsedReq.RequestURI = ""
		s.ParsedReq.URL.Scheme = "http"
		if s.ParsedReq.URL.Port() == "443" {
			s.ParsedReq.URL.Scheme += "s"
		}
		resp, err := s.client.Do(s.ParsedReq)
		if err != nil {
			log.Println("Error making request in proxy >", err)
		} else {
			for k, v := range resp.Header {
				s.Header.add(k, v[0])
			}
			rpBody, readError := io.ReadAll(resp.Body)
			if readError != nil {
				log.Println("Read Error for resp Body >", readError)
			}
			s.ResponseBody = rpBody
		}
	}
}

func (s *HTTPProxyRespHandler) RespHeadOveride() {
	if ovHeaders := s.ProxyHLDQ.GetOVHeaders("res"); len(ovHeaders) > 0 {
		for _, v := range ovHeaders {
			if _, ok := s.Header.hMap[v.Field]; ok {
				if v.Del {
					s.Header.remove(v.Field)
				} else {
					s.Header.add(v.Field, v.Value)
				}
			}
		}
	}
}
