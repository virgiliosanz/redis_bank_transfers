package main

import (
	"flag"
	"fmt"
	"log"

	"iberpay/internal/repository"
	"iberpay/internal/store"
)

type Configuration struct {
	redisURL        string
	prefix          string
	accountsFile    string
	redisMaxRetries int
}

func setUp() *Configuration {
	cfg := Configuration{}
	flag.StringVar(&cfg.redisURL, "u", "redis://127.0.0.1:6379/", "Redis Connection Url")
	flag.StringVar(&cfg.accountsFile, "f", "data/accounts_1000.csv", "File with generated bank accounts.")
	flag.StringVar(&cfg.prefix, "p", "iberpay", "Redis global prefix for application")
	flag.IntVar(&cfg.redisMaxRetries, "r", 100, "Redis max number of retries to adquire lock")

	flag.Parse()

	return &cfg
}

func main() {
	cfg := setUp()
	fmt.Printf("Initializing database on %s\n", cfg.redisURL)
	repo := repository.NewRedisRepository(cfg.redisURL, cfg.prefix, cfg.redisMaxRetries)
	defer repo.Close()

	// delete database data
	if err := repo.FlushDB(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Insert data loaded from: %s\n", cfg.accountsFile)
	// Read accounts from csv
	loader := store.NewCsvLoader(cfg.accountsFile)
	if err := loader.LoadData(); err != nil {
		log.Fatal(err)
	}

	// Initilize the database accounts
	accountStore := store.NewRedisAccountStore(repo)
	loader.InitAccounts(accountStore)
	fmt.Println("Database initialized. Bye!")
}
