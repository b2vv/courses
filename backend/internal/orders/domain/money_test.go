package domain

import (
	"errors"
	"testing"
)

func TestNewMoneyAcceptsPositiveUAHMinorUnits(t *testing.T) {
	money, err := NewMoney("50000", "UAH")
	if err != nil {
		t.Fatalf("NewMoney: %v", err)
	}
	if money.MinorUnits() != MinorUnits(50000) {
		t.Fatalf("minor units=%v want=50000", money.MinorUnits())
	}
	if money.Currency() != CurrencyUAH {
		t.Fatalf("currency=%q want=%q", money.Currency(), CurrencyUAH)
	}
}

func TestNewMoneyRejectsInvalidAmountsAndCurrencies(t *testing.T) {
	for _, test := range []struct {
		name     string
		minor    string
		currency string
	}{
		{name: "negative", minor: "-1", currency: "UAH"},
		{name: "overflow", minor: "9223372036854775808", currency: "UAH"},
		{name: "not a number", minor: "five hundred", currency: "UAH"},
		{name: "unsupported currency", minor: "50000", currency: "EUR"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewMoney(test.minor, test.currency)
			if !errors.Is(err, ErrMoneyInvalid) {
				t.Fatalf("error=%v want=%v", err, ErrMoneyInvalid)
			}
		})
	}
}
