package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

type Response struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

type Server struct {
	mu sync.Mutex

	Response Response
	Routes   map[string]Response

	RequestCount int
	LastMethod   string
	LastPath     string
	LastHeaders  http.Header

	Server *httptest.Server
}

func NewServer(response Response) *Server {
	s := &Server{
		Response: response,
		Routes:   make(map[string]Response),
	}

	s.Server = httptest.NewServer(
		http.HandlerFunc(s.handle),
	)

	return s
}

func NewRouteServer(
	routes map[string]Response,
	defaultResponse Response,
) *Server {
	s := &Server{
		Response: defaultResponse,
		Routes:   make(map[string]Response, len(routes)),
	}

	for path, response := range routes {
		s.Routes[path] = response
	}

	s.Server = httptest.NewServer(
		http.HandlerFunc(s.handle),
	)

	return s
}

func (s *Server) handle(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.mu.Lock()

	s.RequestCount++
	s.LastMethod = r.Method
	s.LastPath = r.URL.Path
	s.LastHeaders = r.Header.Clone()

	response := s.Response

	if routeResponse, ok := s.Routes[r.URL.Path]; ok {
		response = routeResponse
	}

	s.mu.Unlock()

	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(response.StatusCode)

	_, _ = io.WriteString(
		w,
		response.Body,
	)
}

func (s *Server) Close() {
	s.Server.Close()
}

func (s *Server) URL() string {
	return s.Server.URL
}

func (s *Server) Snapshot() (
	requestCount int,
	method string,
	path string,
	headers http.Header,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.RequestCount, s.LastMethod, s.LastPath, s.LastHeaders.Clone()
}
