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
		balance, err := getBalance(ctx, tx, accountKey, category)
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

func getBalance(ctx context.Context, tx *redis.Tx, accountKey, category string) (float64, error) {
	balanceStr, err := tx.HGet(ctx, accountKey, category).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("%s for account: %s, category: %s", ErrAccountNotExists, accountKey, category)
	} else if err != nil {
		return 0, err
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse balance: %v", err)
	}

	return balance, nil
}

func hasSufficientBalance(balance, totalAmount float64) bool {
	return balance >= totalAmount
}
