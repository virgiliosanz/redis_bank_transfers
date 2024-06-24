package store

import (
	"context"

	"redis_bank_transfers/internal/repository"
	"redis_bank_transfers/internal/types"
)

const (
	maxRetries     = 100
	accountPrefix  = ":accounts:"
	bankPrefix     = ":banks:"
	balancesPrefix = ":balances:"
)

type RedisAccountStore struct {
	r   *repository.RedisRepository
	ctx context.Context
}

func NewRedisAccountStore(r *repository.RedisRepository) *RedisAccountStore {
	return &RedisAccountStore{
		r:   r,
		ctx: context.TODO(),
	}
}

func (s *RedisAccountStore) Key(ID string) string {
	return s.r.Prefix + accountPrefix + ID
}

func (s *RedisAccountStore) Set(account types.Account) error {
	err := s.r.Redis.Set(s.ctx, s.Key(account.IBAN), account.Ammount, 0).Err()
	return err
}

func (s *RedisAccountStore) GetBalance(ID string) (int64, error) {
	money, err := s.r.Redis.Get(s.ctx, s.Key(ID)).Int64()
	if err != nil {
		return 0, err
	}
	return money, err
}

// type BankStore interface {
// 	Set(bank string, balance int64) error
// 	GetBalance(ID string) (int64, error)
// 	GetInterBankBalances(ID string) (map[string]int64, error)
// }

type RedisBankStore struct {
	r   *repository.RedisRepository
	ctx context.Context
}

func NewRedisBankStore(r *repository.RedisRepository) *RedisBankStore {
	return &RedisBankStore{
		r:   r,
		ctx: context.TODO(),
	}
}

func (s *RedisBankStore) Key(ID string) string {
	return s.r.Prefix + bankPrefix + ID
}

func (s *RedisBankStore) KeyInterBanks(bankFrom string, bankTo string) string {
	return s.r.Prefix + balancesPrefix + bankFrom + ":" + bankTo
}

func (s *RedisBankStore) Set(bank string, balance int64) error {
	err := s.r.Redis.Set(s.ctx, s.Key(bank), balance, 0).Err()
	return err
}

func (s *RedisBankStore) GetBalance(ID string) (int64, error) {
	money, err := s.r.Redis.Get(s.ctx, s.Key(ID)).Int64()
	if err != nil {
		return 0, err
	}
	return money, err
}

func (s *RedisBankStore) SetInterBankBakance(bankFrom string, bankTo string, balance int64) error {
	key := s.KeyInterBanks(bankFrom, bankTo)
	err := s.r.Redis.Set(s.ctx, key, balance, 0).Err()
	return err
}

func (s *RedisBankStore) GetInterBankBalance(bankFrom string, bankTo string) (int64, error) {
	key := s.KeyInterBanks(bankFrom, bankTo)
	money, err := s.r.Redis.Get(s.ctx, key).Int64()
	if err != nil {
		return 0, err
	}
	return money, err
}
