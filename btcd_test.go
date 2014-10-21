package gochroma_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monetas/btcnet"
	"github.com/monetas/btcrpcclient"
	"github.com/monetas/gochroma"
)

func TestNewBtcdBlockExplorerError(t *testing.T) {
	// Setup
	connConfig := &btcrpcclient.ConnConfig{
		Proxy:        "%gh&%ij",
		HttpPostMode: true,
		DisableTLS:   true,
	}

	// Execute
	_, err := gochroma.NewBtcdBlockExplorer(nil, connConfig)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrConnect)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestBlockCount(t *testing.T) {
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
	blockReaderWriter, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	b := &gochroma.BlockExplorer{blockReaderWriter}
	if err != nil {
		t.Fatal(err)
	}
	// Execute
	count, err := b.BlockCount()
	if err != nil {
		t.Fatal(err)
	}
	// Verify
	if count != countWant {
		t.Fatalf("Did not get back what we expected: got %d, want %d", count, countWant)
	}
}

func TestBlockHash(t *testing.T) {
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
	hash, err := b.BlockHash(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(hash, hashWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", hash, hashWant)
	}
}

func TestBlockHashError(t *testing.T) {
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("nonsense")
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
	_, err = b.BlockHash(1)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestRawBlock(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprint("{\"result\":\"" + bytesStr + "\",\"error\":null,\"id\":1}")
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
	bytesGot, err := b.RawBlock(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestRawBlockError1(t *testing.T) {
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
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
	_, err = b.RawBlock([]byte{0x00})

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrInvalidHash)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestRawBlockError2(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "nonsense")
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
	_, err = b.RawBlock(hash)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestRawTx(t *testing.T) {
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
	bytesGot, err := b.RawTx(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestRawTxError1(t *testing.T) {
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
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
	_, err = b.RawTx([]byte{0x00})

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrInvalidHash)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestRawTxError2(t *testing.T) {
	// Setup
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	hash, _ := hex.DecodeString(hashStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "nonsense")
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
	_, err = b.RawTx(hash)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestTxBlockHash(t *testing.T) {
	// Setup
	txHashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	txHash, _ := hex.DecodeString(txHashStr)
	blockHashStr := "000000000000da63a816b582e5bad4fc3315f709b2ff287980b524b3d16cca22"
	blockHashWant, _ := hex.DecodeString(blockHashStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := "{\"result\": {\"hex\": \"0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000\", \"txid\": \"1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09 f\", \"version\": 1, \"locktime\": 0, \"vin\": [{\"txid\": \"b8bbf855a5a8cc2e9c0726ad499ea8be808b0604971b363050e85f289d0d57aa\", \"vout\": 1, \"scriptSig\": {\"asm\": \"3046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc01037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7fea\", \"hex\": \"493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7fea\"}, \"sequence\": 4294967295}], \"vout\": [{\"value\": 1, \"n\": 0, \"scriptPubKey\": {\"asm\": \"OP_DUP OP_HASH160 9bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa OP_EQUALVERIFY OP_CHECKSIG\", \"hex\": \"76a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac\", \"reqSigs\": 1, \"type\": \"pubkeyhash\", \"addresses\": [\"muiRis7nB1XtfTyKTq4iJzsu6ogeeVDr36\"]}}, {\"value\": 18.63972914, \"n\": 1, \"scriptPubKey\": {\"asm\": \"OP_DUP OP_HASH160 4d273d3a2ce1824d1c6db0764eebb03f368fd9af OP_EQUALVERIFY OP_CHECKSIG\", \"hex\": \"76a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac\", \"reqSigs\": 1, \"type\": \"pubkeyhash\", \"addresses\": [\"mnYuLD9Reoeiwr3fSkzkjiqwHbZG3D2cRd\"]}}], \"blockhash\": \"000000000000da63a816b582e5bad4fc3315f709b2ff287980b524b3d16cca22\", \"confirmations\": 71189, \"time\": 1399048735, \"blocktime\": 1399048735} ,\"error\":null,\"id\":1}"
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
	blockHashGot, err := b.TxBlockHash(txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(blockHashGot, blockHashWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", blockHashGot, blockHashWant)
	}
}

func TestTxBlockHashError2(t *testing.T) {
	txHashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	txHash, _ := hex.DecodeString(txHashStr)
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "nonsense")
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
	_, err = b.TxBlockHash(txHash)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	wantString := "invalid character"
	if !strings.Contains(err.Error(), wantString) {
		t.Fatalf("Got the wrong error, got %v want something with %v", err.Error(), wantString)
	}
}

func TestTxBlockHashError1(t *testing.T) {
	// Setup
	txHashStr := "00"
	txHash, _ := hex.DecodeString(txHashStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
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
	_, err = b.TxBlockHash(txHash)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrInvalidHash)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestMempoolTxs(t *testing.T) {
	// Setup
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	hashWant, _ := hex.DecodeString(hashStr)
	hashRev := make([]byte, len(hashWant))
	copy(hashRev, hashWant)
	for i, j := 0, len(hashRev)-1; i < j; i, j = i+1, j-1 {
		hashRev[i], hashRev[j] = hashRev[j], hashRev[i]
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("{\"result\": [\"%x\"] ,\"error\":null,\"id\":1}", hashRev)
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
	mempoolTxs, err := b.MempoolTxs()
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(mempoolTxs[0], hashWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", mempoolTxs[0], hashWant)
	}
}

func TestMempoolTxsError(t *testing.T) {
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "nonsense")
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
	_, err = b.MempoolTxs()

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

// TestTxOutSpent tests whether a transaction out (aka outpoint) has been
// spent or not. For this test, the response is "null" which corresponds
// to the txout having been spent, hence, true.
func TestTxOutSpent(t *testing.T) {
	// Setup
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	hash, _ := hex.DecodeString(hashStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reply with "null" which is equivalent to the transaction having
		// been spent already.
		response := fmt.Sprintf("{\"result\": null ,\"error\":null,\"id\":1}")
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
	spent, err := b.TxOutSpent(hash, 0, true)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if !*spent {
		t.Fatalf("Did not get back what we expected: want true, got %v", *spent)
	}
}

// TestTxOutSpentError tests the error conditions for TxOutSpent.
func TestTxOutSpentError(t *testing.T) {
	tests := []struct {
		desc    string
		hashStr string
		err     int
	}{
		{
			desc:    "BlockReaderWriter error",
			hashStr: "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f",
			err:     gochroma.ErrBlockRead,
		},
		{
			desc:    "Invalid hash input",
			hashStr: "1d",
			err:     gochroma.ErrInvalidHash,
		},
	}

	for _, test := range tests {

		// Setup
		hash, _ := hex.DecodeString(test.hashStr)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "nonsense")
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
		_, err = b.TxOutSpent(hash, 0, true)

		// Verify
		if err == nil {
			t.Fatal("%v: Got nil where we expected error", test.desc)
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v, wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
		}
	}
}

func TestPublishRawTx(t *testing.T) {
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
	hashGot, err := b.PublishRawTx(bytesToSend)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(hashGot, hashWant) != 0 {
		t.Fatalf("Did not get back what we expected: got %x, want %x", hashGot, hashWant)
	}
}

func TestPublishRawTxError1(t *testing.T) {
	// Setup
	bytesStr := "01"
	bytesToSend, _ := hex.DecodeString(bytesStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
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
	_, err = b.PublishRawTx(bytesToSend)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrInvalidTx)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestPublishRawTxError2(t *testing.T) {
	// Setup
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesToSend, _ := hex.DecodeString(bytesStr)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("nonsense")
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
	_, err = b.PublishRawTx(bytesToSend)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockWrite)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}
