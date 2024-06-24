package service

import "iberpay/internal/types"

type Reporter interface {
	AddEvent(transaction types.Transaction)
	Report()
}
