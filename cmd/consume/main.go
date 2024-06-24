package main

import (
	"flag"
	"fmt"
	"log"

	"iberpay/internal/repository"
	"iberpay/internal/service"
	"iberpay/internal/store"
)

type Configuration struct {
	accountsFile    string
	redisURL        string
	prefix          string
	redisMaxRetries int
}

func setUp() *Configuration {
	cfg := Configuration{}

	flag.StringVar(&cfg.accountsFile, "f", "data/accounts_1000.csv", "File with generated bank accounts.")
	flag.StringVar(&cfg.redisURL, "u", "redis://127.0.0.1:6379/", "Redis Connection Url")
	flag.StringVar(&cfg.prefix, "p", "iberpay", "Redis global prefix for application")
	flag.IntVar(&cfg.redisMaxRetries, "r", 100, "Redis max number of retries to adquire lock")

	flag.Parse()

	return &cfg
}

func main() {
	cfg := setUp()
	fmt.Println("Starting Consuming events using:")
	fmt.Printf("\tAccounts File: %s\n", cfg.accountsFile)
	fmt.Printf("\tAccounts Redis url: %s\n", cfg.redisURL)
	repo := repository.NewRedisRepository(cfg.redisURL, cfg.prefix, cfg.redisMaxRetries)
	defer repo.Close()

	loader := store.NewCsvLoader(cfg.accountsFile)
	if err := loader.LoadData(); err != nil {
		log.Fatal(err)
	}

	streamer := service.NewRedisStream(repo)
	reporter := service.NewFileReporter(repo, *loader.GetAccounts())
	if err := streamer.Consume(reporter); err != nil {
		fmt.Printf("Consume error: %s\n", err)
	}
	reporter.Report()
}
