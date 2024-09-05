package ports

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"transaction-authorizer/internal/core"

	"github.com/go-redis/redis/v8"
)

func NewHTTPServer(ctx context.Context, rdb *redis.Client) {
    http.HandleFunc("/transaction", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        var transaction core.Transaction

        // Decode the JSON payload
        err := json.NewDecoder(r.Body).Decode(&transaction)
        if err != nil {
            log.Printf("Error decoding request: %v", err)
            http.Error(w, "Invalid request payload", http.StatusBadRequest)
            return
        }

        // Handle the transaction using core logic
        err = core.HandleTransaction(ctx, rdb, transaction)
        if err != nil {
            log.Printf("Transaction error: %v", err)
            http.Error(w, "Transaction failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Transaction processed successfully!"))
    })

    log.Println("Server starting on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
