package dto

import (
	"testing"
)

func TestAmount_GetInt(t *testing.T) {
	tests := []struct {
		name   string
		amount Amount
		want   uint64
	}{
		{
			name:   "positive amount",
			amount: Amount(123.45),
			want:   12345,
		},
		{
			name:   "zero amount",
			amount: Amount(0),
			want:   0,
		},
		{
			name:   "negative amount returns zero",
			amount: Amount(-50.00),
			want:   0,
		},
		{
			name:   "small positive amount",
			amount: Amount(0.01),
			want:   1,
		},
		{
			name:   "large amount",
			amount: Amount(999999.99),
			want:   99999999,
		},
		{
			name:   "amount with more than 2 decimal places",
			amount: Amount(10.999),
			want:   1099,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.amount.GetInt(); got != tt.want {
				t.Errorf("Amount.GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAmount_SetAmount(t *testing.T) {
	tests := []struct {
		name       string
		intAmount  uint64
		wantAmount Amount
	}{
		{
			name:       "normal amount",
			intAmount:  12345,
			wantAmount: Amount(123.45),
		},
		{
			name:       "zero amount",
			intAmount:  0,
			wantAmount: Amount(0),
		},
		{
			name:       "small amount",
			intAmount:  1,
			wantAmount: Amount(0.01),
		},
		{
			name:       "large amount",
			intAmount:  99999999,
			wantAmount: Amount(999999.99),
		},
		{
			name:       "round amount",
			intAmount:  100,
			wantAmount: Amount(1.00),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Amount
			a.SetAmount(tt.intAmount)
			if a != tt.wantAmount {
				t.Errorf("Amount.SetAmount() = %v, want %v", a, tt.wantAmount)
			}
		})
	}
}

func TestAmount_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original uint64
	}{
		{
			name:     "small value",
			original: 1,
		},
		{
			name:     "medium value",
			original: 12345,
		},
		{
			name:     "large value",
			original: 99999999,
		},
		{
			name:     "zero value",
			original: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Amount
			a.SetAmount(tt.original)
			got := a.GetInt()
			if got != tt.original {
				t.Errorf("Round trip conversion failed: SetAmount(%v) -> GetInt() = %v", tt.original, got)
			}
		})
	}
}