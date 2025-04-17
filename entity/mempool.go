package entity

import "sync"

type Mempool struct {
	mu  sync.Mutex
	txs []Transaction
}

func NewMempool() *Mempool {
	return &Mempool{txs: []Transaction{}}
}

func (m *Mempool) Add(tx Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs = append(m.txs, tx)
}

func (m *Mempool) All() []Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Transaction(nil), m.txs...)
}
