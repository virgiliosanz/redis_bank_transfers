package store

import (
	"iberpay/internal/types"
)

type AccountStore interface {
	Set(account types.Account) error
	GetBalance(ID string) (int64, error)
}

type BankStore interface {
	Set(bank string, balance int64) error
	GetBalance(ID string) (int64, error)
	SetInterBankBakance(bankFrom string, bankTo string, balance int64) error
	GetInterBankBalance(bankFrom string, bankTo string) (int64, error)
}

type Loader interface {
	LoadData() error
	InitAccounts(store AccountStore) error
	InitBanks(store BankStore) error
	GetAccounts() *map[string]types.Account
	GetBanks() *map[string]string
	GetRandomAccount() types.Account
}
