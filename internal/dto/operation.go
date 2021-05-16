package dto

type Operation struct {
	Wallet      string `json:"wallet"`
	Amount      Amount `json:"amount"`
	Type        string `json:"type"`
	OtherWallet string `json:"other_wallet"`
}
