package service

import (
	"iberpay/internal/types"
)

type Stream interface {
	Send(t types.Transaction) error
	Consume(r Reporter) error
}
