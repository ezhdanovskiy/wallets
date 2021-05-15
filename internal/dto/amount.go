package dto

type Amount float64

func (a *Amount) GetInt() uint64 {
	if a == nil || *a < 0 {
		return 0
	}
	return uint64(*a * 100)
}

func (a *Amount) SetAmount(amount uint64) {
	*a = Amount(amount) / 100
}
