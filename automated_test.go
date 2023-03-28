package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	code := m.Run()
	os.Exit(code)
}

func TestTransactions(t *testing.T) {
	peerList := createAndConnectNPeers(10)

	for i := 0; i < 20; i++ {
		randomTransactions(peerList, 100, 1000, 10)
		time.Sleep(time.Second)
	}

	time.Sleep(5 * time.Second)

	if !checkAgreement(peerList) {
		printTrees(peerList)
		t.Error()
	}

	for _, p := range peerList {
		p.Close()
	}

	printAccounts(peerList)
}

func TestTransactionsTooBig(t *testing.T) {
	peerList := createAndConnectNPeers(10)

	for i := 0; i < 20; i++ {
		randomTransactions(peerList, 100, 1000000, 1)
		time.Sleep(time.Second)
	}

	if !checkAgreement(peerList) {
		t.Error()
	}

	for _, p := range peerList {
		p.Close()
	}

	printAccounts(peerList)
}

func TestTransactionsTooSmall(t *testing.T) {
	peerList := createAndConnectNPeers(10)

	for i := 0; i < 20; i++ {
		from := rand.Intn(len(peerList))
		to := strconv.Itoa(rand.Intn(len(peerList)))
		peerList[from].SendTransaction(to, 0)
		time.Sleep(time.Second)
	}

	if !checkAgreement(peerList) {
		t.Error()
	}

	for _, p := range peerList {
		p.Close()
	}

	printAccounts(peerList)
}

func createAndConnectNPeers(n int) []*peer {
	port := 64000
	peerList := []*peer{}

	// create
	for i := 0; i < n; i++ {
		peerList = append(peerList, createPeer(strconv.Itoa(i)))
	}

	// connect
	for i, p := range peerList {
		if i == 0 {
			p.ConnectAndListen("", getAddress(port))
		} else {
			p.ConnectAndListen(getAddress(rand.Intn(i)+port), getAddress(port+i))
			time.Sleep(time.Duration(100*i/10+1) * time.Millisecond) // presence needs to be sent around
		}
	}

	peerList[0].SendGenesis(peerList...)
	time.Sleep(time.Duration(100*n/10+1) * time.Millisecond)

	return peerList
}

func getAddress(port int) string {
	return ":" + strings.TrimSpace(strconv.Itoa(port))
}

func randomTransactions(peerList []*peer, min int, max int, count int) {
	for i := 0; i < count; i++ {
		from := rand.Intn(len(peerList))
		to := strconv.Itoa(rand.Intn(len(peerList)))
		amount := rand.Intn(max-min) + min
		peerList[from].SendTransaction(to, amount)
	}
}

func checkAgreement(peerList []*peer) bool {
	for _, p := range peerList {
		for k, v := range p.ledger.Accounts {
			if v != peerList[0].ledger.Accounts[k] {
				fmt.Println("bad ledger:", p.info.Alias)
				return false
			}
		}
	}
	return true
}

func printAccounts(peerList []*peer) {
	for _, p := range peerList {
		p.ledger.printAccounts()
	}
}

func printTrees(peerList []*peer) {
	for _, p := range peerList {
		p.PrintTree()
	}
}
