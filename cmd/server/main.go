package main

import (
	"github.com/brecabral/rate-limiter/internal/infra/middleware"
	"github.com/brecabral/rate-limiter/internal/webserver"
)

func main() {
	server := webserver.NewWebServer(":8080")
	limiter := middleware.NewRateLimiterMiddleware()
	handler := limiter.Handle(server.HelloHandler)
	server.AddHandler("/", handler)
	server.Start()
}
