package dto

type Transfer struct {
	WalletFrom string `json:"wallet_from"`
	WalletTo   string `json:"wallet_to"`
	Amount     Amount `json:"amount"`
}
