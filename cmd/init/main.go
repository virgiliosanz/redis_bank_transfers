package main

import (
	"flag"
	"fmt"
	"log"

	"redis_bank_transfers/config"
	"redis_bank_transfers/internal/repository"
	"redis_bank_transfers/internal/store"
)

func setUp() *config.Configuration {
	cfg := config.GetDefaults()
	flag.StringVar(&cfg.RedisURL, "u", cfg.RedisURL, "Redis Connection Url")
	flag.StringVar(&cfg.AccountsFile, "f", cfg.AccountsFile, "File with generated bank accounts.")
	flag.StringVar(&cfg.Prefix, "p", cfg.Prefix, "Redis global prefix for application")
	flag.IntVar(&cfg.RedisMaxRetries, "r", cfg.RedisMaxRetries, "Redis max number of retries to adquire lock")

	flag.Parse()

	return cfg
}

func main() {
	cfg := setUp()
	fmt.Printf("Initializing database on %s\n", cfg.RedisURL)
	repo := repository.NewRedisRepository(cfg.RedisURL, cfg.Prefix, cfg.RedisMaxRetries)
	defer repo.Close()

	// delete database data
	if err := repo.FlushDB(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Insert data loaded from: %s\n", cfg.AccountsFile)
	// Read accounts from csv
	loader := store.NewCsvLoader(cfg.AccountsFile)
	if err := loader.LoadData(); err != nil {
		log.Fatal(err)
	}

	// Initilize the database accounts
	accountStore := store.NewRedisAccountStore(repo)
	loader.InitAccounts(accountStore)
	fmt.Println("Database initialized. Bye!")
}
