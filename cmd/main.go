package main

import (
    "log"
    "transaction-authorizer/internal/adapters"
    "transaction-authorizer/internal/ports"
    "context"
)

func main() {
    ctx := context.Background()

    // Initialize Redis
    rdb := adapters.NewRedisClient("localhost:6379")

    // Start HTTP server
    ports.NewHTTPServer(ctx, rdb)

    log.Println("API is running")
}
