package main

import (
	"crypto/rand"
	"crypto/rsa"
	"dsys/rpc"
	random "math/rand"
	"net"
	"sync"
)

type peer struct {
	info peerInfo

	listener net.Listener
	rpc      *rpc.Rpc

	peerInfoList   []peerInfo
	peerInfoListMu sync.Mutex

	sk                  *rsa.PrivateKey
	blockInfo           blockInfo
	tree                tree
	transactions        map[string]transaction
	transactionsMu      sync.RWMutex
	transactionsQueue   []string
	transactionsQueueMu sync.Mutex

	ledger       *Ledger
	initializing chan struct{}
}

func createPeer(id string) *peer {
	sk, _ := rsa.GenerateKey(rand.Reader, 2048)

	return &peer{
		info: peerInfo{
			Alias: id,
			Pk:    sk.PublicKey,
		},

		sk:                sk,
		transactions:      make(map[string]transaction),
		transactionsQueue: []string{},

		ledger:       MakeLedger(id, sk),
		initializing: make(chan struct{}),
	}
}

type peerInfo struct {
	Alias   string
	Address string
	Pk      rsa.PublicKey
}

func (p *peer) ConnectAndListen(connectAddress string, listenAddress string) {
	p.initialize(listenAddress)

	if connectAddress != "" {
		conn := tryConnect(connectAddress)
		p.rpc.AddConnection(conn)
		p.sendGetPeerInfoList(conn) // Starts listening when received
	} else {
		go p.listenForConnections()
	}

	<-p.initializing
}

func (p *peer) connectToPeers(peerInfoList []peerInfo) {
	p.peerInfoList = append(peerInfoList, p.info)

	startIndex := max(len(peerInfoList)-10, 0)
	// All but the sender since they are already added
	for _, info := range peerInfoList[startIndex : len(peerInfoList)-1] {
		p.rpc.AddConnection(tryConnect(info.Address))
	}

	for _, info := range peerInfoList {
		p.ledger.addAccount(info.Alias, info.Pk)
	}

	p.broadcastPresence(p.info)
}

func (p *peer) listenForConnections() {
	println(p.info.Alias + " is listening on address: " + p.info.Address)
	defer p.listener.Close()

	close(p.initializing)

	for {
		conn, err := p.listener.Accept() // A peer tries to connect to this peer
		if err != nil {
			continue
		}

		p.rpc.AddConnection(conn)
	}
}

func (p *peer) SendGenesis(peerList ...*peer) {
	pks := make([]rsa.PublicKey, len(peerList))
	for i, pk := range peerList {
		pks[i] = pk.sk.PublicKey
	}

	g := genesis{
		Pks:  pks,
		Seed: random.Int(),
	}

	p.BroadcastGenesis(g)
	p.initializeGenesis(g)
}

func (p *peer) SendTransaction(to string, amount int) {
	st := p.createSignedTransaction(to, amount)
	p.addTransaction(st.Transaction)
	p.addToQueue(st.Transaction.ID)
	p.broadcastSignedTransaction(*st)
}

func (p *peer) addToPeerInfoList(info peerInfo) {
	p.peerInfoListMu.Lock()
	defer p.peerInfoListMu.Unlock()
	p.peerInfoList = append(p.peerInfoList, info)
}

func (p *peer) addTransaction(t transaction) {
	p.transactionsMu.Lock()
	defer p.transactionsMu.Unlock()
	p.transactions[t.ID] = t
}

func (p *peer) Close() {
	p.rpc.RemoveAllConnections()
	p.listener.Close()
}

func (p *peer) PrintTree() {
	p.tree.print(p.info.Alias)
}
