package main

type httpServer struct {
	addr string
}

func NewHTTPServer(addr string) *httpServer {
	return &httpServer{addr: addr}
}

func (s *httpServer) Run() error {
	// Implementation to start the HTTP server
	return nil
}