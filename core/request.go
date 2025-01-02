package core

import (
	"net/http"
	"time"
)

type Protocol int

const (
	HTTP Protocol = iota
	WS
)

func (p Protocol) String() string {
	return [...]string{"HTTP", "WS"}[p]
}

type RequestInfo struct {
	Time     time.Duration
	Response *http.Response
	Request  Request
	Err      error
}

type RequestsConfig struct {
	Requests            []Request
	Count_Workers       int
	Delay               time.Duration
	Duration            time.Duration
	RequestChanBufSize  int
	ResponseChanBufSize int
	Secure              bool
	Protocol            Protocol
}

type Request interface {
	GetURI() string
	GetMethod() string
	GetHeaders() http.Header
	GetBody() []byte
}

type HTTPRequest struct {
	*http.Request
	CachedBody []byte
}

func (r *HTTPRequest) GetURI() string {
	return r.URL.String()
}

func (r *HTTPRequest) GetMethod() string {
	return r.Method
}

func (r *HTTPRequest) GetHeaders() http.Header {
	return r.Header
}

func (r *HTTPRequest) GetBody() []byte {
	return r.CachedBody
}

type WSRequest struct {
	URI     string
	Headers http.Header
	Payload []byte
}

func (r *WSRequest) GetURI() string {
	return r.URI
}

func (r *WSRequest) GetMethod() string {
	return "GET"
}

func (r *WSRequest) GetHeaders() http.Header {
	return r.Headers
}

func (r *WSRequest) GetBody() []byte {
	return r.Payload
}

var (
	_ Request = (*HTTPRequest)(nil)
	_ Request = (*WSRequest)(nil)
)
