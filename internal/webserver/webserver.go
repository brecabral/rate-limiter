package webserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/model"
	"github.com/brecabral/rate-limiter/internal/infra/repository"
)

type WebServer struct {
	addr string
	repo repository.StoreKey
}

func NewWebServer(addr string, repo repository.StoreKey) *WebServer {
	return &WebServer{
		addr: addr,
		repo: repo,
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

type CreateKey struct {
	DurationInSeconds  int64 `json:"duration"`
	RateLimitPerSecond int   `json:"rate"`
}

type CreatedKey struct {
	ApiKey string `json:"api-key"`
}

func (s *WebServer) CreateApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var requestKey CreateKey
	if err := json.NewDecoder(r.Body).Decode(&requestKey); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if requestKey.DurationInSeconds <= 0 || requestKey.RateLimitPerSecond <= 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	duration := time.Duration(requestKey.DurationInSeconds) * time.Second
	rate := requestKey.RateLimitPerSecond
	newApiKey := model.CreateApiKey(duration, rate)

	ctx := context.Background()
	if err := s.repo.SaveKey(ctx, newApiKey); err != nil {
		http.Error(w, "Error Creating Token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(CreatedKey{
		ApiKey: newApiKey.Key,
	})
}
