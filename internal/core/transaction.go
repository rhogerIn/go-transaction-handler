package core

import (
    "fmt"
    "strconv"
    "github.com/go-redis/redis/v8"
    "context"
)

type Transaction struct {
    Account     string  `json:"account"`
    TotalAmount float64 `json:"totalAmount"`
    Mcc         string  `json:"mcc"`
    Merchant    string  `json:"merchant"`
}

// HandleTransaction processes the transaction and updates the balance in Redis
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

    balanceStr, err := rdb.HGet(ctx, "account:"+transaction.Account, category).Result()
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

    newBalance := balance - transaction.TotalAmount
    _, err = rdb.HSet(ctx, "account:"+transaction.Account, category, newBalance).Result()
    if err != nil {
        return err
    }

    return nil
}
