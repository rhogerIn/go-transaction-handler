package core

import (
	"context"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestMapMCCToCategory(t *testing.T) {
	tests := []struct {
		mcc			string
		expected	string
	}{
        {MCCFood1, "FOOD"},
        {MCCFood2, "FOOD"},
        {MCCFood3, "MEAL"},
        {MCCFood4, "MEAL"},
        {"1234", "CASH"},
	}
	for _, tt := range tests {
		t.Run(tt.mcc, func(t *testing.T) {
			category := getMCCMapToCategory(tt.mcc)
			if category != tt.expected {
                t.Errorf("got %s, expected %s", category, tt.expected)
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
        t.Fatalf("Could not start mock Redis: %v", err)
	}
	defer s.Close()

    rdb := redis.NewClient(&redis.Options{
        Addr: s.Addr(),
    })
	s.HSet("account:123", "FOOD", "100.0")

    ctx := context.Background()

    balance, err := getBalance(ctx, rdb, "account:123", "FOOD")
    if err != nil {
        t.Fatalf("getBalance returned an error: %v", err)
    }

    if balance != 100.0 {
        t.Errorf("expected balance 100.0, got %.2f", balance)
    }
}

func TestHandleTransaction(t *testing.T) {
    s, err := miniredis.Run()
    if err != nil {
        t.Fatalf("Could not start mock Redis: %v", err)
    }
    defer s.Close()

    rdb := redis.NewClient(&redis.Options{
        Addr: s.Addr(),
    })

    s.HSet("account:123", "FOOD", "100.0")

    ctx := context.Background()

    transaction := Transaction{
        Account:     "123",
        TotalAmount: 50.0,
        Mcc:         "5411",
        Merchant:    "Test Merchant",
    }

    err = HandleTransaction(ctx, rdb, transaction)
    if err != nil {
        t.Fatalf("HandleTransaction returned an error: %v", err)
    }

    balanceStr := s.HGet("account:123", "FOOD")
	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		t.Fatalf("Failed to parse")
	}
    if balance != 50.0 {
        t.Errorf("expected balance 50.0, got %.2f", balance)
    }
}
