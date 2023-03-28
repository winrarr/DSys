package rpc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
)

// Registers a function that the RPC handler will listen for.
// Handler will be called on the payload when the given function is received.
func (r *Rpc) RegisterFunction(function string, handler func(net.Conn, []byte), flooding bool) {
	r.fns[function] = handler
	r.flooding[function] = flooding
}

// Unregisters a function so the RPC handler will no longer listen for it.
func (r *Rpc) UnRegisterFunction(function string) {
	delete(r.fns, function)
}

// Returns a function that handles a received byte array on the specified TCP connection.
func (r *Rpc) Handle(conn net.Conn) chan struct{} {
	stop := make(chan struct{})
	go func(net.Conn) {
		reader := bufio.NewReader(conn)
		for {
			select {
			case <-stop:
				return
			default:
				msg, err := reader.ReadBytes('\n')
				if err != nil {
					return
				}
				r.handleBytes(conn, msg)
			}
		}
	}(conn)
	return stop
}

func (r *Rpc) handleBytes(conn net.Conn, b []byte) {
	fp := bytes.Split(b, []byte(" ")) // function, payload
	r.log("%s received %s", r.alias, fp[0])

	flooding, ok := r.flooding[string(fp[0])]
	if !ok {
		fmt.Printf("received function %s which is not registered", string(fp[0]))
		return
	}
	if flooding {
		if !r.isFlooded(b) {
			r.SendRaw(b, nil, true)
		} else {
			return
		}
	}

	f := r.fns[string(fp[0])]
	f(conn, fp[1][:len(fp[1])-1])
}
