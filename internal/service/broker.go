package service

import "redis_bank_transfers/internal/types"

type Broker interface {
	Transfer(t types.Transaction) types.Transaction
}
