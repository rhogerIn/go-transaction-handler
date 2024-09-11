package core

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
)

type Transaction struct {
	Account     string  `json:"account"`
	TotalAmount float64 `json:"totalAmount"`
	Mcc         string  `json:"mcc"`
	Merchant    string  `json:"merchant"`
}

type TransactionBatch struct {
	Transactions []Transaction
}

const (
	MCCFood1 = "5411"
	MCCFood2 = "5412"
	MCCFood3 = "5811"
	MCCFood4 = "5812"

	ErrAccountNotExists    = "Account or category does not exist"
	ErrInsufficientBalance = "Insufficient balance"
)



func HandleTransaction(ctx context.Context, rdb *redis.Client, transaction Transaction) error {
	category := getMCCMapToCategory(transaction.Mcc)
	if category == "" {
		return fmt.Errorf("invalid MCC: %s", transaction.Mcc)
	}
	accountKey := "account:" + transaction.Account
	err := rdb.Watch(ctx, func(tx *redis.Tx) error {
		balance, err := getBalance(ctx, rdb, accountKey, category)
		if err != nil {
			return err
		}
		if !hasSufficientBalance(balance, transaction.TotalAmount) {
			return fmt.Errorf("%s: available %.2f, required %.2f", ErrInsufficientBalance, balance, transaction.TotalAmount)
		}
		newBalance := balance - transaction.TotalAmount
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, accountKey, category, newBalance)
			return nil
		})
		return err
	}, accountKey)

	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	return nil
}

func getMCCMapToCategory(mcc string) string {
	switch mcc {
	case MCCFood1, MCCFood2:
		return "FOOD"
	case MCCFood3, MCCFood4:
		return "MEAL"
	default:
		return "CASH"
	}
}

func getBalance(ctx context.Context, rdb *redis.Client, accountKey, category string) (float64, error) {
	balanceStr, err := rdb.HGet(ctx, accountKey, category).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("account or category does not exist")
	} else if err != nil {
		return 0, err
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %v", err)
	}

	return balance, nil
}

func GetAccountBalance(ctx context.Context, rdb *redis.Client, account, category string) (float64, error) {
	accountKey := "account:" + account
	balanceStr, err := rdb.HGet(ctx, accountKey, category).Result() // Only catch if is UpperCase Category
	if err == redis.Nil {
		return 0, fmt.Errorf("account or category does not exist")
	} else if err != nil {
		return 0, err
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %v", err)
	}

	return balance, nil
}

func hasSufficientBalance(balance, totalAmount float64) bool {
	return balance >= totalAmount
}

func Deposit(ctx context.Context, rdb *redis.Client, account, category string, amount float64) error {
	accountKey := "account:" + account
	balanceStr, err := rdb.HGet(ctx, accountKey, category).Result() // Only catch if is UpperCase Category
	if err == redis.Nil {
		return fmt.Errorf("account or category does not exist %v", err)
	} else if err != nil {
		return err
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %v", err)
	}

	newBalance := balance + amount
	err = rdb.HSet(ctx, accountKey, category, newBalance).Err()
	if err != nil {
		return err
	}
	return nil
}

func ProcessTransactionBatch(ctx context.Context, rdb *redis.Client, batch TransactionBatch) error {
	pipe := rdb.Pipeline()
	results := make([]*redis.StringCmd, len(batch.Transactions))

	for i, transaction := range batch.Transactions {
		accountKey := "account:" + transaction.Account
		results[i] = pipe.HGet(ctx, accountKey, getMCCMapToCategory(transaction.Mcc))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error executing pipeline: %v", err)
	}

	pipe = rdb.Pipeline()
	for i, transaction := range batch.Transactions {
		balanceStr := results[i].Val()
		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return fmt.Errorf("failed to parse balance: %v", err)
		}

		newBalance := balance - transaction.TotalAmount
		accountKey := "account:" + transaction.Account
		pipe.HSet(ctx, accountKey, getMCCMapToCategory(transaction.Mcc), newBalance)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error executing pipeline: %v", err)
	}

	return nil
}
