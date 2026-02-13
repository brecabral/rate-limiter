package webserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/repository"
	"github.com/brecabral/rate-limiter/internal/infra/token"
)

type WebServer struct {
	addr string
	repo repository.StoreKey
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

type CreateToken struct {
	Duration           int64 `json:"duration"`
	RateLimitPerSecond int   `json:"rate"`
}

type TokenCreated struct {
	ApiKey string `json:"api-key"`
}

func (s *WebServer) POSTCreateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var tokenRequest CreateToken
	if err := json.NewDecoder(r.Body).Decode(&tokenRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if tokenRequest.Duration <= 0 || tokenRequest.RateLimitPerSecond <= 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	duration := time.Duration(tokenRequest.Duration) * time.Second
	rate := tokenRequest.RateLimitPerSecond
	tokenCreated := token.CreateToken(duration, rate)

	if err := s.repo.SaveKey(tokenCreated); err != nil {
		http.Error(w, "Error Creating Token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(TokenCreated{
		ApiKey: tokenCreated.Key,
	})
}
