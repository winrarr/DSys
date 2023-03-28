package main

import (
	"dsys/rpc"
	"fmt"
	"net"
)

// Should have matching send/receive for all RPCs

func (p *peer) sendGetPeerInfoList(conn net.Conn) {
	p.rpc.Send("getPeerInfoList", nil, conn, false)
}

func (p *peer) receivedGetPeerInfoList(conn net.Conn, b []byte) {
	p.sendPeerInfoList(conn)
}

func (p *peer) sendPeerInfoList(conn net.Conn) {
	p.rpc.Send("peerInfoList", p.peerInfoList, conn, false)
}

func (p *peer) receivedPeerInfoList(conn net.Conn, b []byte) {
	// If we already have a peer list
	if len(p.peerInfoList) != 1 {
		return
	}

	var peerInfoList []peerInfo
	tryUnmarshal(b, &peerInfoList)

	p.connectToPeers(peerInfoList)

	go p.listenForConnections()
}

func (p *peer) broadcastPresence(info peerInfo) {
	p.rpc.Send("presence", info, nil, true)
}

func (p *peer) receivedPresence(conn net.Conn, b []byte) {
	var info peerInfo
	tryUnmarshal(b, &info)

	p.addToPeerInfoList(info)
	p.ledger.addAccount(info.Alias, info.Pk)
}

func (p *peer) broadcastSignedTransaction(st signedTransaction) {
	p.rpc.Send("transaction", st, nil, true)
}

func (p *peer) receivedSignedTransaction(conn net.Conn, b []byte) {
	var st signedTransaction
	tryUnmarshal(b, &st)

	t := st.Transaction

	pk, err := decodePk(t.From)
	if err != nil {
		return
	}

	if !verifySignature(pk, st.Signature, t) {
		fmt.Printf("%s could not verify transaction signature...\n", p.info.Alias)
		return
	}

	p.addTransaction(t)
	p.addToQueue(t.ID)
}

func (p *peer) BroadcastGenesis(g genesis) {
	p.rpc.Send("genesis", g, nil, true)
}

func (p *peer) receivedGenesis(conn net.Conn, b []byte) {
	var g genesis
	tryUnmarshal(b, &g)

	p.initializeGenesis(g)
}

func (p *peer) BroadcastBlock(b block) {
	p.rpc.Send("block", b, nil, true)
}

func (p *peer) receivedBlock(conn net.Conn, b []byte) {
	var block block
	tryUnmarshal(b, &block)

	if !p.verifyBlock(block) {
		return
	}

	p.removeFromQueue(block.Transactions...)
	p.tree.insert(block.Slot, block.Transactions, block.Ps, block.Ph)
	p.payWinner(block)
}

func (p *peer) makeRpc() *rpc.Rpc {
	r := rpc.MakeRpc(p.info.Alias, false)
	r.RegisterFunction("presence", p.receivedPresence, true)
	r.RegisterFunction("transaction", p.receivedSignedTransaction, true)
	r.RegisterFunction("getPeerInfoList", p.receivedGetPeerInfoList, false)
	r.RegisterFunction("peerInfoList", p.receivedPeerInfoList, false)
	r.RegisterFunction("genesis", p.receivedGenesis, true)
	r.RegisterFunction("block", p.receivedBlock, true)
	return &r
}
