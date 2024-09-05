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

func HandleTransaction(ctx context.Context, rdb *redis.Client, transaction Transaction) error {
	var category string
	switch transaction.Mcc {
	case "5411", "5412":
		category = "FOOD"
	case "5811", "5812":
		category = "MEAL"
	default:
		category = "CASH"
	}

	accountKey := "account:" + transaction.Account
	err := rdb.Watch(ctx, func(tx *redis.Tx) error {
		balanceStr, err := rdb.HGet(ctx, accountKey, category).Result()
		if err == redis.Nil {
			return fmt.Errorf("account or category does not exist")
		} else if err != nil {
			return err
		}

		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return err
		}

		if balance < transaction.TotalAmount {
			return fmt.Errorf("insufficient balance")
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			newBalance := balance - transaction.TotalAmount
			pipe.HSet(ctx, accountKey, category, newBalance)
			return nil
		})

		return err
	})

	if err != nil {
		return fmt.Errorf("Transaction failed: %v", err)
	}

	return nil
}
