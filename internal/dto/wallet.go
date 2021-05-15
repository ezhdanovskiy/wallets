package dto

type Wallet struct {
	ID      int64  `json:"-"`
	Name    string `json:"name"`
	Balance uint64 `json:"balance"`
}

type CreateWalletRequest struct {
	ID   int64  `json:"-"`
	Name string `json:"name"`
}
