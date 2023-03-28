package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/vishalkuo/bimap"
)

type Ledger struct {
	sk *rsa.PrivateKey

	Accounts   map[string]int
	accountsMu sync.RWMutex
	aliases    *bimap.BiMap

	transactionsDone   map[string]bool
	transactionsDoneMu sync.Mutex
}

func MakeLedger(id string, sk *rsa.PrivateKey) *Ledger {
	l := new(Ledger)
	l.Accounts = make(map[string]int)
	l.aliases = bimap.NewBiMap()
	l.sk = sk
	l.transactionsDone = make(map[string]bool)

	l.addAccount(id, sk.PublicKey)
	return l
}

type transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

func (l *Ledger) createTransaction(to string, amount int) (*transaction, error) {
	pk := l.aliasToPk(to)
	if pk == "" {
		return nil, fmt.Errorf("invalid transaction receiver %v", to)
	}

	return &transaction{
		ID:     uuid.NewString(),
		From:   encodePk(l.sk.PublicKey),
		To:     pk,
		Amount: amount,
	}, nil
}

type signedTransaction struct {
	Transaction transaction
	Signature   []byte
}

func (p *peer) createSignedTransaction(to string, amount int) *signedTransaction {
	t, err := p.ledger.createTransaction(to, amount)
	if err != nil {
		log.Fatal(err)
	}

	return &signedTransaction{
		Transaction: *t,
		Signature:   p.sign(t),
	}
}

func (l *Ledger) transaction(t transaction) {
	l.addTransactionDone(t)

	if t.Amount < 1 {
		fmt.Println("invalid transaction amount")
		return
	}

	//check if sending account becomes negative, but without mutating
	l.accountsMu.RLock()
	if l.Accounts[t.From] < t.Amount {
		fmt.Println("sender will be negative, rejecting transaction")
		// debug.PrintStack()
		return
	}
	l.accountsMu.RUnlock()

	l.transfer(t.From, t.To, t.Amount-1)
}

func (l *Ledger) reverseTransaction(t transaction) {
	t.From, t.To = t.To, t.From

	l.removeTransactionDone(t)

	if t.Amount < 1 {
		fmt.Println("invalid transaction amount")
		return
	}

	//check if sending account becomes negative, but without mutating
	l.accountsMu.RLock()
	if l.Accounts[t.From] < t.Amount {
		fmt.Println("sender will be negative, rejecting transaction...")
		return
	}
	l.accountsMu.RUnlock()

	l.transfer(t.From, t.To, t.Amount+1)
}

func (l *Ledger) transfer(from string, to string, amount int) {
	l.accountsMu.Lock()
	defer l.accountsMu.Unlock()
	l.Accounts[from] -= amount
	l.Accounts[to] += amount
}

func (l *Ledger) addTransactionDone(t transaction) bool {
	l.transactionsDoneMu.Lock()
	defer l.transactionsDoneMu.Unlock()
	if l.transactionsDone[t.ID] {
		return false
	}
	l.transactionsDone[t.ID] = true
	return true
}

func (l *Ledger) removeTransactionDone(t transaction) bool {
	l.transactionsDoneMu.Lock()
	defer l.transactionsDoneMu.Unlock()
	if !l.transactionsDone[t.ID] {
		return false
	}
	l.transactionsDone[t.ID] = false
	return true
}

func (l *Ledger) addAccount(id string, pk rsa.PublicKey) {
	l.accountsMu.Lock()
	defer l.accountsMu.Unlock()
	l.Accounts[encodePk(pk)] = 0
	l.aliases.Insert(id, encodePk(pk))
}

func (l *Ledger) addMoney(pk rsa.PublicKey, amount int) {
	l.accountsMu.Lock()
	defer l.accountsMu.Unlock()
	l.Accounts[encodePk(pk)] += amount
}

func (l *Ledger) getBalance(pk rsa.PublicKey) int {
	l.accountsMu.RLock()
	defer l.accountsMu.RUnlock()
	return l.Accounts[encodePk(pk)]
}

func (l *Ledger) aliasToPk(alias string) string {
	pk, ok := l.aliases.Get(alias)
	if !ok {
		return ""
	}
	return pk.(string)
}

func (l *Ledger) pkToAlias(pk string) string {
	alias, ok := l.aliases.GetInverse(pk)
	if !ok {
		return ""
	}
	return alias.(string)
}

func (l *Ledger) printAccounts() {
	l.accountsMu.RLock()
	defer l.accountsMu.RUnlock()

	// sort accounts
	keys := make([]string, 0)
	for k := range l.Accounts {
		keys = append(keys, l.pkToAlias(k))
	}
	sort.Strings(keys)

	println("---------------------------------------------")
	for _, v := range keys {
		fmt.Printf("%v: %d\n", v, l.Accounts[l.aliasToPk(v)])
	}
	println("---------------------------------------------")
}
