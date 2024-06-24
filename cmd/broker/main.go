package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"

	"redis_bank_transfers/config"
	"redis_bank_transfers/internal/repository"
	"redis_bank_transfers/internal/service"
	"redis_bank_transfers/internal/store"
	"redis_bank_transfers/internal/types"

	"github.com/gosuri/uiprogress"
)

const maxConcurrency = 100

func setUp() *config.Configuration {
	cfg := config.GetDefaults()
	// cfg := config.Configuration{}

	flag.StringVar(&cfg.AccountsFile, "f", cfg.AccountsFile, "File with generated bank accounts.")
	flag.StringVar(&cfg.RedisURL, "u", cfg.RedisURL, "Redis Connection Url")
	flag.StringVar(&cfg.Prefix, "p", cfg.Prefix, "Redis global prefix for application")
	flag.IntVar(&cfg.RedisMaxRetries, "r", cfg.RedisMaxRetries, "Redis max number of retries to adquire lock")
	flag.IntVar(&cfg.Transactions, "t", cfg.Transactions, "Number of transactions per job")
	flag.UintVar(&cfg.TransferMin, "min", cfg.TransferMin, "Minimum ammount of money to trasfer")
	flag.UintVar(&cfg.TransferMax, "max", cfg.TransferMax, "Maximun ammount of money to trasfer")
	flag.UintVar(&cfg.Concurrency, "c", cfg.Concurrency, "Number of Jobs/Threats to work in parallel")

	flag.Parse()

	return cfg
}

func randonMoney(cfg *config.Configuration) int64 {
	max := int(cfg.TransferMax)
	min := int(cfg.TransferMin)
	value := rand.Intn(max-min+1) + min

	return int64(value)
}

func main() {
	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	cfg := setUp()
	fmt.Println("Starting Generating transactions using:")
	fmt.Printf("\tAccounts File: %s\n", cfg.AccountsFile)
	fmt.Printf("\tNumber of transactions: %d\n", cfg.Transactions)
	fmt.Printf("\tNumber of parallel jobs: %d\n", cfg.Concurrency)

	repo := repository.NewRedisRepository(cfg.RedisURL, cfg.Prefix, cfg.RedisMaxRetries)
	defer repo.Close()

	// Read accounts from csv
	loader := store.NewCsvLoader(cfg.AccountsFile)
	if err := loader.LoadData(); err != nil {
		log.Fatal(err)
	}

	// run n threats to do m transactions each
	var wg sync.WaitGroup
	uiprogress.Start()
	bars := [maxConcurrency]*uiprogress.Bar{}
	for i := 1; i <= int(cfg.Concurrency); i++ {
		wg.Add(1)
		bars[i] = uiprogress.AddBar(cfg.Transactions).AppendCompleted().PrependElapsed()
		go func(id int) {
			defer wg.Done()
			// fmt.Fprintf(os.Stderr, "Goroutine %d started\n", id)

			repo := repository.NewRedisRepository(cfg.RedisURL, cfg.Prefix, cfg.RedisMaxRetries)
			defer repo.Close()

			broker := service.NewRedisBroker(repo)
			streamer := service.NewRedisStream(repo)
			// for i := 0; i < cfg.transactions; i++ {
			for bars[id].Incr() {
				t := types.Transaction{
					From:    loader.GetRandomAccount(),
					To:      loader.GetRandomAccount(),
					Ammount: randonMoney(cfg),
				}
				tUpdated := broker.Transfer(t)
				err := streamer.Send(tUpdated)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s", err)
				}
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish
}
