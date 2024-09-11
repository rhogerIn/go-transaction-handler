package ports

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"transaction-authorizer/internal/core"

	"github.com/go-redis/redis/v8"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var validAPIKeys = map[string]bool{
	"valid-api-key": true,
}

func apiKeyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" || !validAPIKeys[apiKey] {
            sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or missing API key")
            return
        }
        next.ServeHTTP(w, r)
    })
}

func sendErrorResponse(w http.ResponseWriter, code int, errorCode string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    errorCode,
		Message: message,
	})
}

func NewHTTPServer(ctx context.Context, rdb *redis.Client) {
	http.Handle("/transaction", apiKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendErrorResponse(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Invalid Request Method")
			return
		}

		var transaction core.Transaction

		err := json.NewDecoder(r.Body).Decode(&transaction)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, "MISSING_PARAMETERS", "Missing account or category")
			return
		}

		err = core.HandleTransaction(ctx, rdb, transaction)
		if err != nil {
			log.Printf("Transaction error: %v", err)
			http.Error(w, "Transaction failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction processed successfully!"))
	})))

	http.Handle("/transaction-pipeline", apiKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendErrorResponse(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Invalid request method")
			return
		}
		var batch core.TransactionBatch
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			sendErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
			return
		}
		err := core.ProcessTransactionBatch(ctx, rdb, batch)
		if err != nil {
			log.Printf("Error processing batch transactions: %v", err)
			sendErrorResponse(w, http.StatusInternalServerError, "BATCH_ERROR", err.Error())
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Batch transactions processed successfully",
		})
	})))

	http.Handle("/balance", apiKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendErrorResponse(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Invalid Request Method")
			return
		}

		account := r.URL.Query().Get("account")
		category := r.URL.Query().Get("category")

		if account == "" || category == "" {
			sendErrorResponse(w, http.StatusBadRequest, "MISSING_PARAMETERS", "Missing account or category")
			return
		}

		balance, err := core.GetAccountBalance(ctx, rdb, account, category)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "BALANCE_ERROR", err.Error())
			return
		}

		json.NewEncoder(w).Encode(map[string]float64{
			"balance": balance,
		})
	})))

	http.Handle("/deposit", apiKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			sendErrorResponse(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Invalid Request Method")
			return
		}

		var req struct {
			Account  string  "account:"
			Category string  "category:"
			Amount   float64 "amount:"
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
			return
		}

		err = core.Deposit(ctx, rdb, req.Account, req.Category, req.Amount)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "DEPOSIT_ERROR", err.Error())
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Successful deposit process",
		})
	})))

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
