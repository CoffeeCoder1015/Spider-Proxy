package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func isUrl(checkUrl string) bool {
	fmt.Println(checkUrl)
	_, err := url.Parse(checkUrl)
	return err == nil
}

type ProxyOut struct {
	client http.Client
}

func (s *ProxyOut) ProxReq(req *http.Request) []byte {
	req.RequestURI = ""
	req.URL.Scheme = "http"
	if req.URL.Port() == "443" {
		req.URL.Scheme += "s"
	}
	resp, err := s.client.Do(req)
	if err != nil {
		log.Println("Response got on ProxyOut", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Request.URL)
	if err != nil {
		log.Println("Read of RespBody on ProxReq", err)
	}
	return respBody
}
