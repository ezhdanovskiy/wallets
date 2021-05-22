package dto

type Wallet struct {
	Name    string `json:"name"`
	Balance uint64 `json:"balance"`
}

type CreateWalletRequest struct {
	Name string `json:"name"`
}
