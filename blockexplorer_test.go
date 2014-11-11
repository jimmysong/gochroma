package gochroma_test

import (
	"bytes"
	"testing"

	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

// NOTE: a lot of useful "constants" are defined in lib_test.go
// these include: blockHash txHash errHash rawBlock normalTx

func TestLatestBlock(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		blockCount: []int64{1},
		blockHash:  [][]byte{blockHash},
		block:      [][]byte{rawBlock},
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
	if bytes.Compare(bytesGot, rawBlock) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, rawBlock)
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
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestRawBlockAtHeight(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		blockHash: [][]byte{blockHash},
		block:     [][]byte{rawBlock},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	bytesGot, err := b.RawBlockAtHeight(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	if bytes.Compare(bytesGot, rawBlock) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, rawBlock)
	}
}

func TestBlockAtHeight(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		blockHash: [][]byte{blockHash},
		block:     [][]byte{rawBlock},
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
	if bytes.Compare(bytesGot, rawBlock) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, rawBlock)
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
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestBlock(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		block: [][]byte{rawBlock},
	}

	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.Block(blockHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, rawBlock) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, rawBlock)
	}
}

func TestPreviousBlock(t *testing.T) {
	blockReaderWriter := &TstBlockReaderWriter{
		block: [][]byte{rawBlock, rawBlock2},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	block, err := b.PreviousBlock(blockHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	bytesGot, err := block.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot, rawBlock2) != 0 {
		t.Fatalf("Did not get block that we expected: got %x, want %x", bytesGot, rawBlock2)
	}
}

func TestTx(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{normalTx},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	tx, err := b.Tx(txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	var bytesGot bytes.Buffer
	err = tx.MsgTx().Serialize(&bytesGot)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(bytesGot.Bytes(), normalTx) != 0 {
		t.Fatalf("Did not get tx that we expected: got %x, want %x", bytesGot.Bytes(), normalTx)
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
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestTxHeight(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash},
		block:       [][]byte{rawBlock},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	height, err := b.TxHeight(txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	heightWant := btcutil.BlockHeightUnknown
	if height != heightWant {
		t.Fatalf("Did not get height that we expected: got %d, want %d", height, heightWant)
	}
}

func TestTxHeightError(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err := b.TxHeight([]byte{0x00})

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
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Errorf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}

func TestOutPointValue(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{normalTx},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
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
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)

	// Execute
	_, err = b.OutPointValue(outPoint)

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

func TestPublishTx(t *testing.T) {
	// Setup
	blockReaderWriter := &TstBlockReaderWriter{
		sendHash: [][]byte{txHash},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(normalTx)
	if err != nil {
		t.Fatalf("couldn't make tx: %v", err)
	}

	// Execute
	shaHash, err := b.PublishTx(tx.MsgTx())
	if err != nil {
		t.Fatal(err)
	}
	hash := gochroma.BigEndianBytes(shaHash)

	// Verify
	if bytes.Compare(hash, txHash) != 0 {
		t.Fatalf("Did not get hash we wanted: got %d, want %d", hash, txHash)
	}
}
