package dto

type Deposit struct {
	Wallet string `json:"wallet"`
	Amount Amount `json:"amount"`
}
