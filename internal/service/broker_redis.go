package service

import (
	"context"
	"fmt"

	"redis_bank_transfers/internal/repository"
	"redis_bank_transfers/internal/store"
	"redis_bank_transfers/internal/types"

	"github.com/redis/go-redis/v9"
)

type RedisBroker struct {
	r   *repository.RedisRepository
	ctx context.Context
	as  *store.RedisAccountStore
	bs  *store.RedisBankStore
}

func NewRedisBroker(repository *repository.RedisRepository) *RedisBroker {
	return &RedisBroker{
		r:   repository,
		ctx: context.Background(),
		as:  store.NewRedisAccountStore(repository),
		bs:  store.NewRedisBankStore(repository),
	}
}

// https://redis.io/docs/latest/develop/interact/transactions/
// # ammount (Amount money to Transfer)
//
// $ WATCH from_account
//
// $ available = GET from_account
//
// ## check if ammount < available
//
// $ MULTI
// $ incrby from_account -ammount
// $ EXEC
// $ incrby to_account ammount
// $ incrby from_bank -ammount
// $ incrby to_bank ammount
// $ incrby fromBank_toBank -ammount
// $ incrby toBank_fromBank ammount
func (b *RedisBroker) Transfer(t types.Transaction) types.Transaction {
	fromAccountKey := b.as.Key(t.From.IBAN)
	toAccountKey := b.as.Key(t.To.IBAN)

	fromBankKey := b.bs.Key(t.From.Bank)
	toBankKey := b.bs.Key(t.To.Bank)

	FromToBankBalanceKey := b.bs.KeyInterBanks(t.From.Bank, t.To.Bank)
	ToFromBankBalanceKey := b.bs.KeyInterBanks(t.To.Bank, t.From.Bank)

	t.Status = types.TransactionRunning
	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Get the current value or zero.
		available, err := tx.Get(b.ctx, fromAccountKey).Int64()
		if err != nil && err != redis.Nil {
			t.Status = types.TransactionError
			t.ErrMsg = err.Error()
			return err
		}

		t.From.Ammount = available

		if available < t.Ammount {
			t.Status = types.TransactionCancelled
			t.ErrMsg = fmt.Sprintf("Not enough money in the account %s (%d) to send: %d",
				t.From.IBAN, t.From.Ammount, t.Ammount)
			return fmt.Errorf("%w: %d", types.ErrNotEnoughMoney, available)
		}

		// Actual operation (local in optimistic lock).
		// Operation is commited only if the watched keys remain unchanged.
		_, err = tx.TxPipelined(b.ctx, func(pipe redis.Pipeliner) error {
			pipe.IncrBy(b.ctx, fromAccountKey, -t.Ammount)
			return nil
		})
		return err
	}

	// Retry if the key has been changed.
	var err error
	var i int
	for i = 0; i < b.r.MaxRetries; i++ {
		err = b.r.Redis.Watch(b.ctx, txf, fromAccountKey)
		if err == nil {
			// Success!!!! Increment the rest of the values atomically
			if t.From.Bank != t.To.Bank { // Si es entre el mismo banco no actualizamos
				if err = b.r.Redis.IncrBy(b.ctx, fromBankKey, -t.Ammount).Err(); err != nil {
					t.Status = types.TransactionError
					t.ErrMsg = fmt.Sprintf("Error in transaction: taking money out from the source bank :%s", err.Error())
					return t
				}
				if err = b.r.Redis.IncrBy(b.ctx, toBankKey, t.Ammount).Err(); err != nil {
					t.Status = types.TransactionError
					t.ErrMsg = fmt.Sprintf("Error in transaction: sending money to the target bank :%s", err.Error())
					return t
				}

				if err = b.r.Redis.IncrBy(b.ctx, FromToBankBalanceKey, t.Ammount).Err(); err != nil {
					t.Status = types.TransactionError
					t.ErrMsg = fmt.Sprintf("Error in transaction: Updating balance from-to :%s", err.Error())
					return t
				}

				if err = b.r.Redis.IncrBy(b.ctx, ToFromBankBalanceKey, -t.Ammount).Err(); err != nil {
					t.Status = types.TransactionError
					t.ErrMsg = fmt.Sprintf("Error in transaction: Updating balance to-from: %s", err.Error())
					return t
				}

			}
			if err = b.r.Redis.IncrBy(b.ctx, toAccountKey, t.Ammount).Err(); err != nil {
				t.Status = types.TransactionError
				t.ErrMsg = fmt.Sprintf("Error in transaction: sending money to the target account :%s", err.Error())
				return t
			}

			t.Status = types.TransactionCompleted
			t.ErrMsg = ""
			return t
		}

		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}

		return t
	}

	t.Status = types.TransactionError
	t.ErrMsg = fmt.Sprintf("Error in transaction -> %d tries: %s", i, err.Error())
	return t
}
