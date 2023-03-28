package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"math/big"
	"time"
)

type blockInfo struct {
	slot int
	seed int
}

type genesis struct {
	Pks  []rsa.PublicKey
	Seed int
}

type block struct {
	Transactions []string
	Pk           *rsa.PublicKey
	Ps           int
	Ph           []byte
	Slot         int
	Draw         []byte
	Shb          []byte
}

func (p *peer) startSendingBlocks() {
	for {
		p.blockInfo.slot++
		time.Sleep(time.Second)
		p.nextSlot()
	}
}

func (p *peer) nextSlot() {
	draw := p.computeDraw()
	value := p.computeValue(draw, p.sk.PublicKey)
	if !aboveHardness(value) {
		return
	}

	n := p.tree.insertNext(p.blockInfo.slot, p.clearQueue())
	if n.Slot == n.parent.Slot {
		return
	}

	b := block{
		Transactions: n.Transactions,
		Pk:           &p.sk.PublicKey,
		Ps:           n.parent.Slot,
		Ph:           n.parent.hash(),
		Slot:         p.blockInfo.slot,
		Draw:         draw,
	}

	b.Shb = p.sign(b.Ph, b.Transactions)

	p.BroadcastBlock(b)
	p.payWinner(b)
}

func (p *peer) payWinner(b block) {
	amount := 10 + len(b.Transactions)
	p.ledger.addMoney(*b.Pk, amount)
}

func (p *peer) computeDraw() []byte {
	bytes := p.sign("LOTTERY", p.blockInfo.seed, p.blockInfo.slot)
	return bytes
}

func (p *peer) computeValue(draw []byte, pk rsa.PublicKey) *big.Int {
	drawHash := sha256.Sum256(draw)
	drawHashValue := new(big.Int).SetBytes(drawHash[:])
	a := big.NewInt(int64(p.ledger.getBalance(pk)))
	return new(big.Int).Mul(drawHashValue, a)
}

func aboveHardness(value *big.Int) bool {
	if value.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	v, _ := new(big.Int).SetString(value.String()[:3], 10)
	return v.Cmp(big.NewInt(900)) == 1
}

func (p *peer) verifyBlock(b block) bool {
	if !verifySignature(b.Pk, b.Draw, "LOTTERY", p.blockInfo.seed, b.Slot) {
		return false
	}

	if !verifySignature(b.Pk, b.Shb, b.Ph, b.Transactions) {
		return false
	}

	if !aboveHardness(p.computeValue(b.Draw, *b.Pk)) {
		return false
	}

	return true
}

func (p *peer) addToQueue(id string) {
	p.transactionsQueueMu.Lock()
	defer p.transactionsQueueMu.Unlock()
	p.transactionsQueue = append(p.transactionsQueue, id)
}

func (p *peer) removeFromQueue(ids ...string) {
	p.transactionsQueueMu.Lock()
	defer p.transactionsQueueMu.Unlock()
	for _, t1 := range ids {
		for i, t2 := range p.transactionsQueue {
			if t1 == t2 {
				if i == len(p.transactionsQueue)-1 {
					p.transactionsQueue = p.transactionsQueue[:i]
				} else {
					p.transactionsQueue[i] = p.transactionsQueue[len(p.transactionsQueue)-1]
				}
			}
		}
	}
}

func (p *peer) clearQueue() []string {
	p.transactionsQueueMu.Lock()
	defer p.transactionsQueueMu.Unlock()
	temp := p.transactionsQueue
	p.transactionsQueue = []string{}
	return temp
}

func (p *peer) runBlock(n *node) {
	p.transactionsMu.RLock()
	defer p.transactionsMu.RUnlock()
	for _, s := range n.Transactions {
		p.ledger.transaction(p.transactions[s])
	}
}

func (p *peer) undoBlock(n *node) {
	p.transactionsMu.RLock()
	defer p.transactionsMu.RUnlock()
	for _, s := range n.Transactions {
		p.ledger.reverseTransaction(p.transactions[s])
	}
}
