package rpc

import (
	"fmt"
	"net"
	"sync"
)

type Rpc struct {
	conns map[net.Conn](chan struct{})
	fns   map[string]func(net.Conn, []byte)

	flooded   map[string]bool
	floodedMu sync.RWMutex
	flooding  map[string]bool

	logging bool
	alias   string
}

// Returns an RPC handler that you can add functions to.
func MakeRpc(alias string, log bool) Rpc {
	return Rpc{
		conns: make(map[net.Conn]chan struct{}),
		fns:   make(map[string]func(net.Conn, []byte)),

		flooded:  make(map[string]bool),
		flooding: make(map[string]bool),

		logging: log,
		alias:   alias,
	}
}

// Starts listening on a connection
func (r *Rpc) AddConnection(conn net.Conn) {
	r.conns[conn] = r.Handle(conn)
}

// Stops listening on a connection
func (r *Rpc) RemoveConnection(conn net.Conn) {
	close(r.conns[conn])
	delete(r.conns, conn)
	conn.Close()
}

func (r *Rpc) RemoveAllConnections() {
	for conn, c := range r.conns {
		close(c)
		delete(r.conns, conn)
		conn.Close()
	}
}

func (r *Rpc) isFlooded(b []byte) bool {
	r.floodedMu.RLock()
	defer r.floodedMu.RUnlock()
	return r.flooded[string(b)]
}

func (r *Rpc) addFlooded(b []byte) {
	r.floodedMu.Lock()
	defer r.floodedMu.Unlock()
	r.flooded[string(b)] = true
}

func (r *Rpc) log(s string, args ...interface{}) {
	if !r.logging {
		return
	}
	fmt.Printf(s+"\n", args...)
}
