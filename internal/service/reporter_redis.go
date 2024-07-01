package service

import (
	"fmt"
	"strings"

	"redis_bank_transfers/internal/repository"
	"redis_bank_transfers/internal/store"
	"redis_bank_transfers/internal/types"

	"github.com/gosuri/uilive"
)

type stats struct {
	Accounts            map[string]types.Account
	Banks               map[string]int64
	Balances            map[string]int64
	TransacionErrors    int64
	TransacionCompleted int64
	TransacionCancelled int64
	Errors              []string
}

func (s *stats) addData(transaction types.Transaction) {
	switch transaction.Status {
	case types.TransactionCompleted:
		s.TransacionCompleted++

		var tmp types.Account
		tmp = s.Accounts[transaction.From.IBAN]
		tmp.Ammount -= transaction.Ammount
		s.Accounts[transaction.From.IBAN] = tmp

		tmp = s.Accounts[transaction.To.IBAN]
		tmp.Ammount += transaction.Ammount
		s.Accounts[transaction.To.IBAN] = tmp

		var bank string
		var current int64
		bank = s.Accounts[transaction.From.IBAN].Bank
		current = s.Banks[bank]
		current -= transaction.Ammount
		s.Banks[bank] = current

		bank = s.Accounts[transaction.To.IBAN].Bank
		current = s.Banks[bank]
		current += transaction.Ammount
		s.Banks[bank] = current

		var balanceKey string
		balanceKey = transaction.From.Bank + ":" + transaction.To.Bank
		current = s.Balances[balanceKey]
		current += transaction.Ammount
		s.Balances[balanceKey] = current

		balanceKey = transaction.To.Bank + ":" + transaction.From.Bank
		current = s.Balances[balanceKey]
		current -= transaction.Ammount
		s.Balances[balanceKey] = current

	case types.TransactionCreated:
		msg := "Incomplete!! Transaction in Creation State: " + transaction.ErrMsg
		s.Errors = append(s.Errors, msg)
		s.TransacionErrors++

	case types.TransactionRunning:
		msg := "Incomplete!! Transaction in Running State: " + transaction.ErrMsg
		s.Errors = append(s.Errors, msg)
		s.TransacionErrors++

	case types.TransactionCancelled:
		// Transaction was cancelled because the origin account didn't have enough
		// money to send the ammount requested
		s.TransacionCancelled++

	case types.TransactionError:
		// Any other error while working on the transaction. Most common is that the
		// proccess could not obtain the "lock" to read and modify the ammount in
		// a loop of 100 tries
		s.Errors = append(s.Errors, transaction.ErrMsg)
		s.TransacionErrors++

	default:
		msg := "Unknown Error: " + transaction.ErrMsg
		s.Errors = append(s.Errors, msg)
		s.TransacionErrors++
	}
}

func newStats(accounts map[string]types.Account) *stats {
	return &stats{
		Accounts:            accounts,
		Banks:               make(map[string]int64),
		Balances:            make(map[string]int64),
		TransacionErrors:    0,
		TransacionCompleted: 0,
		TransacionCancelled: 0,
	}
}

type FileReporter struct {
	repo  *repository.RedisRepository
	stats *stats
	w     *uilive.Writer
}

func NewFileReporter(repo *repository.RedisRepository, accounts map[string]types.Account) *FileReporter {
	r := &FileReporter{
		repo:  repo,
		stats: newStats(accounts),
		w:     uilive.New(),
	}
	r.w.Start()
	return r
}

func (r *FileReporter) AddEvent(transaction types.Transaction) {
	r.stats.addData(transaction)
	// fmt.Fprintf(r.w, "Transactions Completed: %d Cancelled: %d Errors: %d\n",
	// 	r.stats.TransacionCompleted, r.stats.TransacionCancelled, r.stats.TransacionErrors)
}

func (r *FileReporter) checkAccountBalances() {
	fmt.Printf("Checking Account balances")
	store := store.NewRedisAccountStore(r.repo)
	Ok := "....Ok"
	for IBAN, account := range r.stats.Accounts {
		balance, _ := store.GetBalance(IBAN)
		if balance != account.Ammount {
			fmt.Printf("%s: %d != %d\n", IBAN, balance, account.Ammount)
			Ok = "........No OK"
		}
	}

	fmt.Println(Ok)
}

func (r *FileReporter) checkBankBalances() {
	fmt.Printf("Checking Bank balances")
	store := store.NewRedisBankStore(r.repo)
	Ok := "....Ok"
	var total int64 = 0
	for bank, ammount := range r.stats.Banks {
		balance, _ := store.GetBalance(bank)
		if balance != ammount {
			fmt.Printf("%s: %d != %d\n", bank, balance, ammount)
			Ok = "........No OK"
		}
		total += ammount
	}
	fmt.Println(Ok)
	fmt.Printf("The sun of balances should be 0:  %d\n", total)
}

func (r *FileReporter) checkInterBankBalances() {
	fmt.Printf("Checking inter Bank balances")
	store := store.NewRedisBankStore(r.repo)
	Ok := "....Ok"
	var total int64 = 0
	for interBanks, ammount := range r.stats.Balances {
		banks := strings.Split(interBanks, ":")
		balance, _ := store.GetInterBankBalance(banks[0], banks[1])

		if balance != ammount {
			fmt.Printf("%s: %d != %d\n", interBanks, balance, ammount)
			Ok = "........No OK"
		}
		total += ammount
	}

	fmt.Println(Ok)
	fmt.Printf("The sum of inter bank balances should be 0:  %d\n", total)
}

func (r *FileReporter) Report() {
	r.checkAccountBalances()
	r.checkBankBalances()
	r.checkInterBankBalances()
	r.w.Stop()
	fmt.Printf("\nTransactions Completed: %d Cancelled: %d Errors: %d\n",
		r.stats.TransacionCompleted, r.stats.TransacionCancelled, r.stats.TransacionErrors)

	fmt.Println("Errors found in the stream events:")
	for i := 0; i < len(r.stats.Errors); i++ {
		fmt.Println(r.stats.Errors[i])
	}
}
