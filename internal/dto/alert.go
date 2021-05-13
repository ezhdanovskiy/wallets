package dto

type Wallet struct {
	ID   int64  `json:"-"`
	Name string `json:"name"`
}
