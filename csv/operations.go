// Package csv contains methods for generating csv.
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/ezhdanovskiy/wallets/internal/dto"
)

// ConvertOperations converts operations to csv format.
func ConvertOperations(operations []dto.Operation) ([]byte, error) {
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)

	err := csvWriter.Write([]string{"wallet", "amount", "type", "other_wallet", "timestamp"})
	if err != nil {
		return nil, err
	}

	for _, op := range operations {
		err := csvWriter.Write([]string{
			op.Wallet,
			fmt.Sprintf("%v", op.Amount),
			op.Type,
			op.OtherWallet,
			fmt.Sprintf("%v", op.Timestamp),
		})
		if err != nil {
			return nil, err
		}
	}

	csvWriter.Flush()
	return buf.Bytes(), nil
}
