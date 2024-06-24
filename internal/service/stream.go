package service

import (
	"redis_bank_transfers/internal/types"
)

type Stream interface {
	Send(t types.Transaction) error
	Consume(r Reporter) error
}
