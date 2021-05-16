package repository

import (
	"time"
)

type Wallet struct {
	Name      string    `db:"name"`
	Balance   uint64    `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Operation struct {
	ID          int64     `db:"id"`
	Wallet      string    `db:"wallet"`
	Type        string    `db:"type"`
	Amount      uint64    `db:"amount"`
	OtherWallet string    `db:"other_wallet"`
	CreatedAt   time.Time `db:"created_at"`
}
