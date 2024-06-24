package service

import "iberpay/internal/types"

type Broker interface {
	Transfer(t types.Transaction) types.Transaction
}
