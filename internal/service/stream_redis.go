package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"iberpay/internal/repository"
	"iberpay/internal/types"

	"github.com/redis/go-redis/v9"
)

type RedisStream struct {
	repo               *repository.RedisRepository
	ctx                context.Context
	transactionsStream string
}

func NewRedisStream(repo *repository.RedisRepository) *RedisStream {
	return &RedisStream{
		repo:               repo,
		ctx:                context.Background(),
		transactionsStream: repo.Prefix + ":transactions",
	}
}

func (s *RedisStream) Send(t types.Transaction) error {
	// values := map[string]interface{}{
	// 	"account_from": t.From.IBAN,
	// 	"account_to":   t.To.IBAN,
	// 	"ammount":      t.Ammount,
	// 	"disposable":   t.From.Ammount,
	// 	"status":       t.Status,
	// }
	msg, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("error marshaling transacion to json: %w", err)
	}

	values := map[string]string{"json": string(msg[:])}
	err = s.repo.Redis.XAdd(s.ctx, &redis.XAddArgs{
		Stream: s.transactionsStream,
		Values: values,
	}).Err()
	if err != nil {
		return fmt.Errorf("error sending event: %w", err)
	}

	return nil
}

func (s *RedisStream) Consume(r Reporter) error {
	lastCompletedID := "0"
	args := &redis.XReadArgs{Block: 2 * time.Second}
	transaction := types.Transaction{}
	for {
		args.Streams = []string{s.transactionsStream, lastCompletedID}
		streams, err := s.repo.Redis.XRead(s.ctx, args).Result()
		if err != nil {
			// log.Printf("No new messages from queue: %s", err)
			// panic(err)
			break
		}

		for _, message := range streams[0].Messages {
			lastCompletedID = message.ID
			value := message.Values["json"].(string)
			err := json.Unmarshal([]byte(value), &transaction)
			if err != nil {
				log.Println(value)
				log.Printf("Error in json: %s", err)
				continue
			}

			r.AddEvent(transaction)
		}
	}
	return nil
}
