package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"

	"iberpay/internal/repository"
	"iberpay/internal/service"
	"iberpay/internal/store"
	"iberpay/internal/types"

	"github.com/gosuri/uiprogress"
)

const maxConcurrency = 100

type Configuration struct {
	accountsFile    string
	redisURL        string
	prefix          string
	redisMaxRetries int
	transactions    int
	transferMin     uint
	transferMax     uint
	concurrency     uint
}

func setUp() *Configuration {
	cfg := Configuration{}

	flag.StringVar(&cfg.accountsFile, "f", "data/accounts_1000.csv", "File with generated bank accounts.")
	flag.StringVar(&cfg.redisURL, "u", "redis://127.0.0.1:6379/", "Redis Connection Url")
	flag.StringVar(&cfg.prefix, "p", "iberpay", "Redis global prefix for application")
	flag.IntVar(&cfg.redisMaxRetries, "r", 100, "Redis max number of retries to adquire lock")
	flag.IntVar(&cfg.transactions, "t", 10000, "Number of transactions per job")
	flag.UintVar(&cfg.transferMin, "min", 100, "Minimum ammount of money to trasfer")
	flag.UintVar(&cfg.transferMax, "max", 1000, "Maximun ammount of money to trasfer")
	flag.UintVar(&cfg.concurrency, "c", 20, "Number of Jobs/Threats to work in parallel")

	flag.Parse()

	return &cfg
}

func randonMoney(cfg *Configuration) int64 {
	max := int(cfg.transferMax)
	min := int(cfg.transferMin)
	value := rand.Intn(max-min+1) + min

	return int64(value)
}

func main() {
	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	cfg := setUp()
	fmt.Println("Starting Generating transactions using:")
	fmt.Printf("\tAccounts File: %s\n", cfg.accountsFile)
	fmt.Printf("\tNumber of transactions: %d\n", cfg.transactions)
	fmt.Printf("\tNumber of parallel jobs: %d\n", cfg.concurrency)

	repo := repository.NewRedisRepository(cfg.redisURL, cfg.prefix, cfg.redisMaxRetries)
	defer repo.Close()

	// Read accounts from csv
	loader := store.NewCsvLoader(cfg.accountsFile)
	if err := loader.LoadData(); err != nil {
		log.Fatal(err)
	}

	// run n threats to do m transactions each
	var wg sync.WaitGroup
	uiprogress.Start()
	bars := [maxConcurrency]*uiprogress.Bar{}
	for i := 1; i <= int(cfg.concurrency); i++ {
		wg.Add(1)
		bars[i] = uiprogress.AddBar(cfg.transactions).AppendCompleted().PrependElapsed()
		go func(id int) {
			defer wg.Done()
			// fmt.Fprintf(os.Stderr, "Goroutine %d started\n", id)

			repo := repository.NewRedisRepository(cfg.redisURL, cfg.prefix, cfg.redisMaxRetries)
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
