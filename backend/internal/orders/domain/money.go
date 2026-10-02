package domain

import "strconv"

type Currency string

const CurrencyUAH Currency = "UAH"

// MinorUnits keeps persisted monetary amounts strongly typed at package boundaries.
type MinorUnits int64

type Money struct {
	minor    MinorUnits
	currency Currency
}

func NewMoney(rawMinor, rawCurrency string) (Money, error) {
	minor, err := strconv.ParseInt(rawMinor, 10, 64)
	if err != nil || minor < 0 || Currency(rawCurrency) != CurrencyUAH {
		return Money{}, ErrMoneyInvalid
	}
	return Money{minor: MinorUnits(minor), currency: CurrencyUAH}, nil
}

func (money Money) MinorUnits() MinorUnits { return money.minor }

func (money Money) Currency() Currency { return money.currency }

func (money Money) Equal(other Money) bool {
	return money.minor == other.minor && money.currency == other.currency
}
