package gochroma_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conformal/btclog"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/jimmysong/gochroma"
)

var log btclog.Logger

func TestGetBlockCount(t *testing.T) {
	// Setup
	countWant := int64(47834)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		response := fmt.Sprintf("{\"result\":%d,\"error\":null,\"id\":1}", countWant)
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()
	net := &btcnet.TestNet3Params
	connConfig := &btcrpcclient.ConnConfig{
		Host:         ts.URL[7:],
		HttpPostMode: true,
		DisableTLS:   true,
	}
	b, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	if err != nil {
		t.Fatal(err)
	}
	// Execute
	count, err := b.GetBlockCount()
	if err != nil {
		t.Fatal(err)
	}
	// Verify
	if count != countWant {
		t.Fatalf("Did not get back what we expected: got %d, want %d", count, countWant)
	}
}

func TestGetBlockHash(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hashWant, _ := hex.DecodeString(hashStr)
	hashRev := make([]byte, len(hashWant))
	copy(hashRev, hashWant)
	for i, j := 0, len(hashRev)-1; i < j; i, j = i+1, j-1 {
		hashRev[i], hashRev[j] = hashRev[j], hashRev[i]
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("{\"result\":\"%x\",\"error\":null,\"id\":1}", hashRev)
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()
	net := &btcnet.TestNet3Params
	connConfig := &btcrpcclient.ConnConfig{
		Host:         ts.URL[7:],
		HttpPostMode: true,
		DisableTLS:   true,
	}
	b, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Execute
	hash, err := b.GetBlockHash(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(hash, hashWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", hash, hashWant)
	}
}

func TestGetBlock(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000096e88000000000000000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprint("{\"result\":\"\",\"error\":null,\"id\":1}")
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()
	net := &btcnet.TestNet3Params
	connConfig := &btcrpcclient.ConnConfig{
		Host:         ts.URL[7:],
		HttpPostMode: true,
		DisableTLS:   true,
	}
	b, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Execute
	block, err := b.GetBlock(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestGetRawTransaction(t *testing.T) {
	// Setup
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := "{\"result\":\"" + bytesStr + "\",\"error\":null,\"id\":1}"
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()
	net := &btcnet.TestNet3Params
	connConfig := &btcrpcclient.ConnConfig{
		Host:         ts.URL[7:],
		HttpPostMode: true,
		DisableTLS:   true,
	}
	b, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Execute
	bytesGot, err := b.GetRawTransaction(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestSendRawTransaction(t *testing.T) {
	// Setup
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	hashWant, _ := hex.DecodeString(hashStr)
	hashRev := make([]byte, len(hashWant))
	copy(hashRev, hashWant)
	for i, j := 0, len(hashRev)-1; i < j; i, j = i+1, j-1 {
		hashRev[i], hashRev[j] = hashRev[j], hashRev[i]
	}
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesToSend, _ := hex.DecodeString(bytesStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("{\"result\":\"%x\",\"error\":null,\"id\":1}", hashRev)
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()
	net := &btcnet.TestNet3Params
	connConfig := &btcrpcclient.ConnConfig{
		Host:         ts.URL[7:],
		HttpPostMode: true,
		DisableTLS:   true,
	}
	b, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Execute
	hashGot, err := b.SendRawTransaction(bytesToSend)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(hashGot, hashWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", hashGot, hashWant)
	}
}
