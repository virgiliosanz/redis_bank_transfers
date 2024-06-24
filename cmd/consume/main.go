package main

import (
	"flag"
	"fmt"
	"log"

	"redis_bank_transfers/config"
	"redis_bank_transfers/internal/repository"
	"redis_bank_transfers/internal/service"
	"redis_bank_transfers/internal/store"
)

func setUp() *config.Configuration {
	cfg := config.GetDefaults()

	flag.StringVar(&cfg.AccountsFile, "f", cfg.AccountsFile, "File with generated bank accounts.")
	flag.StringVar(&cfg.RedisURL, "u", cfg.RedisURL, "Redis Connection Url")
	flag.StringVar(&cfg.Prefix, "p", cfg.Prefix, "Redis global prefix for application")
	flag.IntVar(&cfg.RedisMaxRetries, "r", cfg.RedisMaxRetries, "Redis max number of retries to adquire lock")

	flag.Parse()

	return cfg
}

func main() {
	cfg := setUp()
	fmt.Println("Starting Consuming events using:")
	fmt.Printf("\tAccounts File: %s\n", cfg.AccountsFile)
	fmt.Printf("\tAccounts Redis url: %s\n", cfg.RedisURL)
	repo := repository.NewRedisRepository(cfg.RedisURL, cfg.Prefix, cfg.RedisMaxRetries)
	defer repo.Close()

	loader := store.NewCsvLoader(cfg.AccountsFile)
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
