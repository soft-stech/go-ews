package ews

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

// https://docs.microsoft.com/en-us/exchange/client-developer/web-service-reference/ews-operations-in-exchange

const (
	soapStart = `<?xml version="1.0" encoding="utf-8" ?>
	<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages" 
		xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types" 
		xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  		<soap:Header>
    		<t:RequestServerVersion Version="Exchange2010_SP1" />
  		</soap:Header>
  		<soap:Body>
`
	soapEnd = `</soap:Body></soap:Envelope>`
)

type Config struct {
	Dump bool
}

type Client struct {
	EWSAddr  string
	Username string
	Password string
	config   *Config
}

func NewClient(ewsAddr, username, password string) *Client {
	return NewClientWithConfig(ewsAddr, username, password,
		&Config{Dump: false},
	)
}

func NewClientWithConfig(ewsAddr, username, password string, config *Config) *Client {
	return &Client{
		EWSAddr:  ewsAddr,
		Username: username,
		Password: password,
		config:   config,
	}
}

func (c *Client) sendAndReceive(body []byte) ([]byte, error) {

	bb := []byte(soapStart)
	bb = append(bb, body...)
	bb = append(bb, soapEnd...)

	req, err := http.NewRequest("POST", c.EWSAddr, bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	logRequest(c, req)

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", "text/xml")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logResponse(c, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, NewSoapError(resp)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, err
}

func logRequest(c *Client, req *http.Request) {
	if c.config != nil && c.config.Dump {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Request:\n%v\n----\n", string(dump))
	}
}

func logResponse(c *Client, resp *http.Response) {
	if c.config != nil && c.config.Dump {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Response:\n%v\n----\n", string(dump))
	}
}