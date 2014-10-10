package gochroma_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

func TestLatestBlock(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		blockCount: []int64{1},
		blockHash:  [][]byte{hash},
		block:      [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.LatestBlock()
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestLatestBlockError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err := b.LatestBlock()

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
}

func TestRawBlockAtHeight(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		blockHash: [][]byte{hash},
		block:     [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	bytesGot, err := b.RawBlockAtHeight(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestBlockAtHeight(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		blockHash: [][]byte{hash},
		block:     [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.BlockAtHeight(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestBlockAtHeightError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err := b.BlockAtHeight(1)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
}

func TestBlock(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		block: [][]byte{bytesWant},
	}

	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.Block(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestPreviousBlock(t *testing.T) {
	// Setup
	hashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr1 := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesCurrent, _ := hex.DecodeString(bytesStr1)
	bytesStr2 := "020000000548c8eb8c91c25c598f7bcb7e3d2f2f14971836c5796bb1023d1d0000000000836b81f78a4421c6bf663353fba5cf2a53d8ee3f76e4f47e96784e1ab1f3803dbee12e54c0ff3f1b995ba51c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2503047b04184b6e434d696e657242519dceb367fae996d0542ee1be2b7d0000000009020000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr2)
	blockReaderWriter := &TstBlockReaderWriter{
		block: [][]byte{bytesCurrent, bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.PreviousBlock(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestTx(t *testing.T) {
	// Setup
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	hash, _ := hex.DecodeString(hashStr)
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	tx, err := b.Tx(hash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	var bytesGot bytes.Buffer
	err = tx.MsgTx().Serialize(&bytesGot)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot.Bytes(), bytesWant) != 0 {
		t.Fatalf("Did not get tx that we expected: got %x, want %x", bytesGot.Bytes(), bytesWant)
	}
}

func TestTxError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err := b.Tx([]byte{0x00})

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
}

func TestOutPointValue(t *testing.T) {
	// Setup
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	shaHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", hashStr, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)

	// Execute
	value, err := b.OutPointValue(outPoint)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	wantValue := int64(100000000)
	if value != wantValue {
		t.Fatalf("Did not get value that we expected: got %d, want %d", value, wantValue)
	}
}

func TestOutPointValueError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	hashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	shaHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", hashStr, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)

	// Execute
	_, err = b.OutPointValue(outPoint)

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
}

func TestTxBlock(t *testing.T) {
	// Setup
	txHashStr := "1d235c4ea39e7f3151e784283319485f4b5eb92e553ee6d307c0201b4125e09f"
	txHash, _ := hex.DecodeString(txHashStr)
	blockHashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	blockHash, _ := hex.DecodeString(blockHashStr)
	bytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash},
		block:       [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.TxBlock(txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, bytesWant) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, bytesWant)
	}
}

func TestTxBlockError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err := b.TxBlock([]byte{0x00})

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
}

func TestPreviousBlockError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err := b.PreviousBlock([]byte{0x00})

	// Verify
	if err == nil {
		t.Fatal("Got nil where we expected error")
	}
}
