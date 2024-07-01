package config

type Configuration struct {
	AccountsFile    string
	RedisURL        string
	Prefix          string
	RedisMaxRetries int
	Transactions    int
	TransferMin     uint
	TransferMax     uint
	Concurrency     uint
}

func GetDefaults() *Configuration {
	return &Configuration{
		AccountsFile: "data/accounts_100.csv",
		RedisURL:     "redis://127.0.0.1:6379/",
		// RedisURL:        "redis://127.0.0.1:12000/",
		Prefix:          "banks_transfers",
		RedisMaxRetries: 100,
		Transactions:    5000,
		TransferMin:     100,
		TransferMax:     5000,
		Concurrency:     20,
	}
}
