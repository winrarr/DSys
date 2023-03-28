package rpc

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
)

// Sends the specified bytes to the given connection. If the given connection
// is nil, sends the specified bytes to all added connections.
func (r *Rpc) Send(function string, payload interface{}, conn net.Conn, flood bool) {
	r.SendRaw(encodeMessage(function, payload), conn, flood)
}

func (r *Rpc) SendRaw(b []byte, conn net.Conn, flood bool) {
	if flood {
		r.addFlooded(b)
	}
	fp := bytes.Split(b, []byte(" ")) // function, payload
	r.log("%s sending %s", r.alias, fp[0])

	if conn != nil {
		_, err := conn.Write(b)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		for conn := range r.conns {
			_, err := conn.Write(b)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Encodes the given function with the given payload to be sent as an RPC.
func encodeMessage(function string, payload interface{}) []byte {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	return append(append([]byte(function+" "), payloadBytes...), '\n')
}
