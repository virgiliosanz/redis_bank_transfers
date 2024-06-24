package store

import (
	"encoding/csv"
	"math/rand"
	"os"
	"strconv"

	"redis_bank_transfers/internal/types"

	"golang.org/x/exp/maps"
)

type CsvLoader struct {
	accounts map[string]types.Account
	fileName string
}

func NewCsvLoader(accountsFileName string) *CsvLoader {
	return &CsvLoader{
		accounts: make(map[string]types.Account),
		fileName: accountsFileName,
	}
}

func readCSV(fname string, skipHeader bool) ([][]string, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if skipHeader {
		if _, err = reader.Read(); err != nil {
			return nil, err
		}
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (l *CsvLoader) LoadData() error {
	records, err := readCSV(l.fileName, true)
	if err != nil {
		return err
	}

	var ammount int
	for _, r := range records {
		ammount, _ = strconv.Atoi(r[3])
		l.accounts[r[0]] = types.Account{
			IBAN:    r[0],
			Bank:    r[1],
			Person:  r[2],
			Ammount: int64(ammount),
		}
	}
	return nil
}

func (l *CsvLoader) InitAccounts(store AccountStore) error {
	for _, account := range l.accounts {
		if err := store.Set(account); err != nil {
			return err
		}
	}
	return nil
}

func (l *CsvLoader) GetRandomAccount() types.Account {
	// keys := reflect.ValueOf(l.accounts).MapKeys()
	// ret := keys[rand.Intn(len(keys))].String()

	keys := maps.Keys(l.accounts)
	ret := keys[rand.Intn(len(keys))]

	return l.accounts[ret]
}

func (l *CsvLoader) InitBanks(store BankStore) error {
	// TODO: Implement InitBanks
	return nil
}

func (l *CsvLoader) GetAccounts() *map[string]types.Account {
	return &l.accounts
}

func (l *CsvLoader) GetBanks() *map[string]string {
	// TODO: Implement GetBanks
	return nil
}
