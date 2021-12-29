package main

import (
	"fmt"
	"net/http"
	"net/url"
)

/*
Basic design
HTTP Proxy - Wrapped HTTP Server
- Handles request
	- reads request URL *3 url redirect
	- redo request on itself to Read URL
		(1)- coppies and sends initial request (headers,etc) *1 - request data overide
	- responds with response got @ (2) *2 - response data overide
Feature impl points -
*1 - replace data in headers with custom ones
*2 - replace data in headers with custom ones
*3 - replace url with custom one
		- can create virtual url
*/

func isUrl(checkUrl string) bool {
	fmt.Println(checkUrl)
	_, err := url.Parse(checkUrl)
	return err == nil
}

type ProxyOut struct {
	client http.Client
}

/*
func (s *ProxyOut) ProxReq(req *http.Request) ([]byte, error) {
	req.RequestURI = ""
	req.URL.Scheme = "http"
	if req.URL.Port() == "443" {
		req.URL.Scheme += "s"
	}
	resp, err := s.client.Do(req)
	if err != nil {
		log.Println("Response got on ProxyOut", err)
	} else {
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		fmt.Println(resp.Request.URL)
		if err != nil {
			log.Println("Read of RespBody on ProxReq", err)
		}
		return respBody, nil
	}
	return []byte{}, CreateError("Failed to make request")
}

func ProxyResponse(req *http.Request) ([]byte, error) {
	if !isUrl(req.RequestURI) {
		return []byte{}, CreateError("Not a URL")
	}
	rDat, err := ProxReq(req)
	if err != nil {
		return []byte{}, err
	}
	return rDat, nil
}
*/
