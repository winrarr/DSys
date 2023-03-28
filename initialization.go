package main

import (
	"log"
	"net"
)

func (p *peer) initialize(listenAddress string) {
	p.initializeListener(listenAddress)
	p.initializeTree()
	p.initializeRPC()
}

func (p *peer) initializeListener(listenAddress string) {
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	p.listener = ln
	p.info.Address = getLocalAddress(p.listener)
	p.addToPeerInfoList(p.info)
}

func (p *peer) initializeTree() {
	p.tree = makeTree(p.runBlock, p.undoBlock)
}

func (p *peer) initializeRPC() {
	p.rpc = p.makeRpc()
}

func (p *peer) initializeGenesis(g genesis) {
	p.blockInfo.seed = g.Seed
	for _, pk := range g.Pks {
		p.ledger.addMoney(pk, 1000000)
	}
	go p.startSendingBlocks()
}
