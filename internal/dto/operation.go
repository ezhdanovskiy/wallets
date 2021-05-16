package dto

import (
	"time"
)

type Operation struct {
	Wallet      string    `json:"wallet"`
	Amount      Amount    `json:"amount"`
	Type        string    `json:"type"`
	OtherWallet string    `json:"other_wallet"`
	Timestamp   time.Time `json:"timestamp"`
}

type OperationsFilter struct {
	Wallet    string
	Type      string
	StartDate int64
	EndDate   int64
	Limit     int64
	Offset    int64
	Format    string // todo: json/csv
}
