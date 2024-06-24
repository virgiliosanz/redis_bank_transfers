package service

import "redis_bank_transfers/internal/types"

type Reporter interface {
	AddEvent(transaction types.Transaction)
	Report()
}
