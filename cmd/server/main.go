package main

import (
	"github.com/brecabral/rate-limiter/internal/infra/limiter"
	"github.com/brecabral/rate-limiter/internal/infra/middleware"
	"github.com/brecabral/rate-limiter/internal/webserver"
)

func main() {
	server := webserver.NewWebServer(":8080")
	limiter := limiter.NewRateLimiter()
	limiterMiddleware := middleware.NewRateLimiterMiddleware(limiter)
	handler := limiterMiddleware.Handle(server.HelloHandler)
	server.AddHandler("/", handler)
	server.AddHandler("/token", server.POSTCreateToken)
	server.Start()
}
