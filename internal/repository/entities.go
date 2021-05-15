package repository

import (
	"time"
)

type Wallet struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Balance   uint64    `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
