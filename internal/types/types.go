package types

import (
	"errors"
)

var (
	ErrNotImplemented    = errors.New("method not implemented")
	ErrMaxRetriesReached = errors.New("max number of retries")
	ErrNotEnoughMoney    = errors.New("not enough money in account")
)

type Account struct {
	IBAN    string `json:"iban"`
	Bank    string `json:"bank"`
	Person  string `json:"person"`
	Ammount int64  `json:"ammount"`
}

const (
	TransactionCreated = iota
	TransactionRunning
	TransactionCompleted
	TransactionCancelled
	TransactionError
)

type Transaction struct {
	ErrMsg  string  `json:"err"`
	From    Account `json:"from"`
	To      Account `json:"to"`
	Ammount int64   `json:"ammount"`
	Status  uint    `json:"status"`
}
