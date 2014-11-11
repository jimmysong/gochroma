package gochroma_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

func TestCreateWallet(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)

	// execute
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)

	// validate
	if err != nil {
		t.Fatal(err)
	}
	err = wallet.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestOpenWallet(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	err = wallet.Close()
	if err != nil {
		t.Fatal(err)
	}

	// execute
	wallet2, err := gochroma.OpenWallet(tmpLocation)
	if err != nil {
		t.Fatal(err)
	}
	err = wallet.Close()
	if err != nil {
		t.Fatal(err)
	}
	if wallet.Net != wallet2.Net {
		t.Fatal("expected nets to be the same")
	}
}

func TestNewColorId(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()

	for i := 1; i < 700; i++ {
		// execute
		colorId, err := wallet.NewColorId()

		// validate
		if err != nil {
			t.Fatal(err)
		}
		if colorId == nil {
			t.Fatalf("%d: expected something got nothing", i)
		}
		if int(colorId[0]) != i%256 {
			t.Fatalf("incorrect colorId: got %d, want %d", int(colorId[0]), i)
		}
	}
}

func TestAddDefinition(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	colorStr := "SPOBC:00000000000000000000000000000000:0:1"

	// execute
	id, err := wallet.FetchOrAddDefinition(colorStr)

	// validate
	if err != nil {
		t.Fatal(err)
	}
	if id[0] != 1 {
		t.Fatal("expected id to be 1")
	}
}

func TestNewAddress(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	colorStr := "SPOBC:00000000000000000000000000000000:0:1"
	_, err = wallet.FetchOrAddDefinition(colorStr)
	if err != nil {
		t.Fatal(err)
	}
	cd, err := gochroma.NewColorDefinitionFromStr(colorStr)
	if err != nil {
		t.Fatal(err)
	}

	// execute
	uncoloredAddr, err := wallet.NewUncoloredAddress()
	if err != nil {
		t.Fatal(err)
	}
	issuingAddr, err := wallet.NewIssuingAddress()
	if err != nil {
		t.Fatal(err)
	}
	colorAddr, err := wallet.NewColorAddress(cd)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	want := "SkADtnrbXupRMYVdhjfxYjBM8eD9zisuKC"
	got := uncoloredAddr.String()
	if got != want {
		t.Fatalf("address differs from expected: want %v, got %v", want, got)
	}
	want = "SRMqaNp3AARBQdy12tHe4KcvLLnSfhk161"
	got = issuingAddr.String()
	if got != want {
		t.Fatalf("address differs from expected: want %v, got %v", want, got)
	}
	want = "SYQikix7fWMMQ9LPFtrNKwVzTDaNUzVYvN"
	got = colorAddr.String()
	if got != want {
		t.Fatalf("address differs from expected: want %v, got %v", want, got)
	}
}

func TestFetchOutPointId(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)

	// execute
	outPointId, err := wallet.FetchOutPointId(outPoint)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	if outPointId != nil {
		t.Fatalf("expected a nil outpoint Id, got %v", *outPointId)
	}

}

func TestNewUncoloredOutPoint(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{normalTx},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)

	// execute
	cop, err := wallet.NewUncoloredOutPoint(b, outPoint)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	value := cop.ColorValue
	wantValue := gochroma.ColorValue(100000000)
	if value != wantValue {
		t.Fatalf("Did not get value that we expected: got %d, want %d", value, wantValue)
	}
	outPointGot, err := cop.OutPoint()
	if err != nil {
		t.Fatal(err)
	}
	if outPointGot.Index != outPoint.Index {
		t.Fatalf("Did not get value that we expected: got %d, want %d", outPointGot.Index, outPoint.Index)
	}
}

func TestNewColorOutPoint(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash},
		block:       [][]byte{rawBlock},
		rawTx:       [][]byte{genesisTx, genesisTx},
		txOutSpents: []bool{false},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)
	shaBytes := gochroma.BigEndianBytes(tx.Sha())
	txString := fmt.Sprintf("%x", shaBytes)
	colorStr := "SPOBC:" + txString + ":0:1"
	cd, err := gochroma.NewColorDefinitionFromStr(colorStr)
	if err != nil {
		t.Fatal(err)
	}

	// execute
	cop, err := wallet.NewColorOutPoint(b, outPoint, cd)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	wantValue := gochroma.ColorValue(1)
	if cop.ColorValue != wantValue {
		t.Fatalf("Did not get value that we expected: got %d, want %d", cop.ColorValue, wantValue)
	}
	cin, err := cop.ColorIn()
	if err != nil {
		t.Fatal(err)
	}
	if cin.ColorValue != wantValue {
		t.Fatalf("Did not get value that we expected: got %d, want %d", cin.ColorValue, wantValue)
	}
}

func TestIssueColor(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	blockReaderWriter := &TstBlockReaderWriter{
		block:       [][]byte{rawBlock},
		txBlockHash: [][]byte{blockHash},
		rawTx:       [][]byte{specialTx, specialTx, specialTx, specialTx, specialTx, specialTx, specialTx},
		txOutSpents: []bool{false, false, false},
		sendHash:    [][]byte{txHash},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)
	_, err = wallet.NewUncoloredOutPoint(b, outPoint)
	if err != nil {
		t.Fatal(err)
	}
	kernelCode := "SPOBC"
	kernel, err := gochroma.GetColorKernel(kernelCode)
	if err != nil {
		t.Fatal(err)
	}
	value := gochroma.ColorValue(1)

	// execute
	cd, err := wallet.IssueColor(b, kernel, value, int64(10000))
	if err != nil {
		t.Fatal(err)
	}

	// validate
	gotStr := cd.HashString()
	wantStr := fmt.Sprintf("%v:%x:0", kernelCode, txHash)
	if gotStr != wantStr {
		t.Fatalf("color definition different than expected: want %v, got %v",
			wantStr, gotStr)
	}
}

func TestAllColors(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	num := 10
	for i := 0; i < num; i++ {
		def := fmt.Sprintf("SPOBC:%x:0:1", seed)
		_, err = wallet.FetchOrAddDefinition(def)
		seed[0] += 1
	}

	// execute
	cds, err := wallet.AllColors()
	if err != nil {
		t.Fatal(err)
	}

	// validate
	if len(cds) != num {
		t.Fatalf("different number of color defs: want %d, got %d", num, len(cds))
	}
}

func TestColorBalance(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash},
		block:       [][]byte{rawBlock},
		rawTx:       [][]byte{genesisTx, genesisTx},
		txOutSpents: []bool{false},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)
	shaBytes := gochroma.BigEndianBytes(tx.Sha())
	txString := fmt.Sprintf("%x", shaBytes)
	colorStr := "SPOBC:" + txString + ":0:1"
	cd, err := gochroma.NewColorDefinitionFromStr(colorStr)
	if err != nil {
		t.Fatal(err)
	}
	cid, err := wallet.FetchOrAddDefinition(colorStr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = wallet.NewColorOutPoint(b, outPoint, cd)
	if err != nil {
		t.Fatal(err)
	}

	// execute
	balance, err := wallet.ColorBalance(cid)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	want := gochroma.ColorValue(1)
	if *balance != want {
		t.Fatalf("balance not what we wanted: want %v, got %v", want, balance)
	}
}

func TestSend(t *testing.T) {
	// setup
	systemTmp := os.TempDir()
	defer os.RemoveAll(systemTmp)
	tmpPath, err := ioutil.TempDir(systemTmp, "wallet_test")
	tmpLocation := filepath.Join(tmpPath, "wallet.bin")
	net := &btcnet.SimNetParams
	seed := make([]byte, 32)
	wallet, err := gochroma.CreateWallet(seed, tmpLocation, net)
	if err != nil {
		t.Fatal(err)
	}
	defer wallet.Close()
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash},
		block:       [][]byte{rawBlock},
		rawTx:       [][]byte{specialTx, specialTx, specialTx, niceTx, niceTx, niceTx, niceTx, niceTx, niceTx, niceTx, niceTx, niceTx, niceTx},
		txOutSpents: []bool{false, false, false, false, false, false, false, false},
		sendHash:    [][]byte{txHash, txHash},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)
	_, err = wallet.NewUncoloredOutPoint(b, outPoint)
	if err != nil {
		t.Fatal(err)
	}
	shaHash, err = gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	outPoint = btcwire.NewOutPoint(shaHash, 0)
	_, err = wallet.NewUncoloredOutPoint(b, outPoint)
	if err != nil {
		t.Fatal(err)
	}
	kernelCode := "SPOBC"
	kernel, err := gochroma.GetColorKernel(kernelCode)
	if err != nil {
		t.Fatal(err)
	}
	value := gochroma.ColorValue(1)
	cd, err := wallet.IssueColor(b, kernel, value, int64(10000))
	if err != nil {
		t.Fatal(err)
	}
	addr, err := wallet.NewColorAddress(cd)
	if err != nil {
		t.Fatal(err)
	}
	addrMap := map[btcutil.Address]gochroma.ColorValue{addr: value}
	fee := int64(100)
	// execute
	tx, err := wallet.Send(b, cd, addrMap, fee)

	// validate
	if err != nil {
		t.Fatal(err)
	}
	if len(tx.TxIn) != 2 {
		t.Fatalf("expected more inputs: want %d, got %d", 2, len(tx.TxIn))
	}
	if len(tx.TxOut) != 2 {
		t.Fatalf("expected more outputs: want %d, got %d", 2, len(tx.TxOut))
	}
	spobc := kernel.(*gochroma.SPOBC)
	want := spobc.MinimumSatoshi
	if tx.TxOut[0].Value != want {
		t.Fatalf("unexpected output at 0: want %d, got %d", want, tx.TxOut[0].Value)
	}
	want = 20000 - want - fee
	if tx.TxOut[1].Value != want {
		t.Fatalf("unexpected output at 1: want %d, got %d", want, tx.TxOut[1].Value)
	}
}
