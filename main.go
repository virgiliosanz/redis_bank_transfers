package main

import (
	"flag"
	"fmt"
	"os"
)

// banks:           "data/banks.csv",
// accounts:        "data/accounts_5000.csv",
// redisURL:        "redis://127.0.0.1:6379/",
// prefix:          "iberpay",
// transactions:    1000,
// transferMin:     100,
// transferMax:     1000,
// concurrency:     10,
// redisMaxRetries: 100,

func main() {
	redisUrl := flag.String("url", "redis://127.0.0.1:6379/", "Redis Connection Url")
	accountsFile := flag.String("af", "data/accounts_1000.csv", "File with generated bank accounts.")
	appPrefix := flag.String("prefix", "my_app", "Redis global prefix for application")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	brokerCmd := flag.NewFlagSet("broker", flag.ExitOnError)
	brokerThreats := brokerCmd.Int("t", 8, "Number of parallel jobs/threats to run")
	brokerN := brokerCmd.Int("n", 1000, "Number of transaction per job/threat")
	brokerTransferMin := brokerCmd.Int("transfer_min", 100, "Number of transaction per job/threat")
	brokerTransferMax := brokerCmd.Int("transfer_max", 1000, "Number of transaction per job/threat")

	consumeCmd := flag.NewFlagSet("consume", flag.ExitOnError)
	consumeErrorOrTransaction := consumeCmd.Bool("e", false, "Consume errors if true, other wise consume transactions")

	// noComandMsg := "expected 'init', 'broker' or 'consume' subcommands"
	// if len(os.Args) < 2 {
	// 	fmt.Println(noComandMsg)
	// 	os.Exit(1)
	// }

	fmt.Println("Global params:")
	fmt.Printf("\tRedis Url: %s\n", *redisUrl)
	fmt.Printf("\tAccounts File: %s\n", *accountsFile)
	fmt.Printf("\tApp Prefix: %s\n", *appPrefix)

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'init'")

	case "broker":
		brokerCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'broker':")
		fmt.Printf("\tNumber of Threats/Jobs: %d\n", *brokerThreats)
		fmt.Printf("\tNumber of transactions per Job: %d\n", *brokerN)
		fmt.Printf("\tMinimun money to transfer: %d\n", *brokerTransferMin)
		fmt.Printf("\tMaximun money to transfer: %d\n", *brokerTransferMax)

	case "consume":
		consumeCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'consume'")
		fmt.Printf("\tConsume Errors: %t\n", *consumeErrorOrTransaction)
	}
}
