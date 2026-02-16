package main

import (
	"log"
	"os"
	"strconv"

	"github.com/brecabral/rate-limiter/internal/infra/limiter"
	"github.com/brecabral/rate-limiter/internal/infra/middleware"
	"github.com/brecabral/rate-limiter/internal/infra/repository"
	"github.com/brecabral/rate-limiter/internal/webserver"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	maxRequestsIP, err := strconv.Atoi(os.Getenv("MAX_REQUESTS_BY_IP_PER_SECOND"))
	if err != nil {
		maxRequestsIP = 10
	}
	blockTime, err := strconv.Atoi(os.Getenv("BLOCK_TIME_IN_SECONDS"))
	if err != nil {
		blockTime = 60
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redisDB = 0
	}

	repo := repository.NewRedisRepository(redisAddr, redisPassword, redisDB)
	server := webserver.NewWebServer(":8080", repo)
	limiter := limiter.NewRateLimiter(repo, maxRequestsIP, blockTime)

	limiterMiddleware := middleware.NewRateLimiterMiddleware(limiter)
	handler := limiterMiddleware.Handle(server.HelloHandler)

	server.AddHandler("/", handler)
	server.AddHandler("/api-key", server.CreateApiKeyHandler)
	server.Start()
}
