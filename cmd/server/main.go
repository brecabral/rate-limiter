package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brecabral/rate-limiter/internal/infra/limiter"
	"github.com/brecabral/rate-limiter/internal/infra/middleware"
	"github.com/brecabral/rate-limiter/internal/infra/model"
	"github.com/brecabral/rate-limiter/internal/infra/repository"
	"github.com/brecabral/rate-limiter/internal/webserver"
	"github.com/joho/godotenv"
)

const TEST_KEY = "teste_key"

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

	repo := repository.NewRedisRepository()
	testToken := model.CreateManualToken(TEST_KEY, time.Hour, 10)
	repo.SaveKey(testToken)
	server := webserver.NewWebServer(":8080", repo)
	limiter := limiter.NewRateLimiter(repo, maxRequestsIP, blockTime)

	limiterMiddleware := middleware.NewRateLimiterMiddleware(limiter)
	handler := limiterMiddleware.Handle(server.HelloHandler)

	server.AddHandler("/", handler)
	server.AddHandler("/token", server.POSTCreateToken)
	server.Start()
}
