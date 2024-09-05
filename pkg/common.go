package pkg

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
)

func ConvertToCents(v float64) int64 {
	return int64(v * 100)
}

func ConvertToUnit(v int64) float64 {
	f, _ := decimal.NewFromInt(v).DivRound(decimal.NewFromInt(100), 2).Float64()
	return f
}

func CreateAccountNumber() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1e9))
	return fmt.Sprintf("%09d", n.Int64())
}
