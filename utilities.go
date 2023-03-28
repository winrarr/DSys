package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net"
	"reflect"
	"strconv"
	"strings"
)

//manual implementation of max for integers
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

//A way to get our IPv4 address
func getLocalAddress(ln net.Listener) string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		println(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ip := localAddr.IP.String()

	lnAddr := ln.Addr().String()
	_, port, _ := net.SplitHostPort(lnAddr)

	return strings.TrimSpace(ip + ":" + port)
}

//encode public key to string
func encodePk(pk rsa.PublicKey) string {
	return pk.N.String() + "," + strconv.Itoa(pk.E)
}

//decode string of a public key to its actual public key
func decodePk(pk string) (*rsa.PublicKey, error) {
	ne := strings.Split(pk, ",")
	if len(ne) != 2 {
		return nil, errors.New("bad public key encoding")
	}

	n, ok := new(big.Int).SetString(ne[0], 10)
	if !ok {
		return nil, errors.New("bad encoding of n")
	}

	e, err := strconv.Atoi(ne[1])
	if err != nil {
		return nil, errors.New("bad encoding of e")
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

func tryConnect(address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func tryUnmarshal(b []byte, v interface{}) {
	err := json.Unmarshal(b, &v)
	if err != nil {
		log.Fatalf("%v: %v", reflect.TypeOf(v), err)
	}
	try(json.Unmarshal(b, &v))
}

// Returns the signature of the given data
func (p *peer) sign(data ...interface{}) []byte {
	signature, err := rsa.SignPKCS1v15(rand.Reader, p.sk, crypto.SHA256, hashObject(data))
	if err != nil {
		log.Fatal(err)
	}

	return signature
}

func verifySignature(pk *rsa.PublicKey, signature []byte, data ...interface{}) bool {
	return rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashObject(data), signature) == nil
}

func hashObject(v ...interface{}) []byte {
	hash := sha256.Sum256(objectToBytes(v))
	return hash[:]
}

func objectToBytes(v ...interface{}) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}
