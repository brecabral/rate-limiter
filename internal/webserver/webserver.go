package webserver

import "net/http"

type WebServer struct {
	addr string
}

func NewWebServer(addr string) *WebServer {
	return &WebServer{
		addr: addr,
	}
}

func (s *WebServer) Start() {
	http.ListenAndServe(s.addr, nil)
}

func (s *WebServer) AddHandler(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, handler)
}

func (s *WebServer) HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World"))
}
