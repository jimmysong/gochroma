package gochroma_test

import (
	"bytes"
	"testing"

	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

func TestSerializeUint32(t *testing.T) {
	wantInt := uint32(14783)
	s := gochroma.SerializeUint32(wantInt)
	gotInt := gochroma.DeserializeUint32(s)

	if gotInt != wantInt {
		t.Fatalf("didn't get what we expected, want %v, got %v", wantInt, gotInt)
	}
}

func TestSerializeOutPoint(t *testing.T) {
	// setup
	shaHash, err := gochroma.NewShaHash(txHash)
	if err != nil {
		t.Fatalf("failed to convert hash %v: %v", txHash, err)
	}
	op := btcwire.NewOutPoint(shaHash, 0)

	// execute
	s, err := gochroma.SerializeOutPoint(op)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	dop, err := gochroma.DeserializeOutPoint(s)
	if err != nil {
		t.Fatal(err)
	}
	if !dop.Hash.IsEqual(&op.Hash) {
		t.Fatalf("didn't get what we expected: want %v, got %v", op.Hash, dop.Hash)
	}
	if dop.Index != op.Index {
		t.Fatalf("didn't get what we expected: want %v, got %v", op.Index, dop.Index)
	}

}

func TestSerializeColorOutPoint(t *testing.T) {
	// setup
	wantId := []byte{5, 1, 2, 1}
	cop := &gochroma.ColorOutPoint{
		Id:            wantId,
		Tx:            txHash,
		Index:         1,
		Value:         100000,
		Color:         wantId,
		ColorValue:    500,
		SpendingTx:    txHash,
		SpendingIndex: 2,
		PkScript:      txHash,
	}

	// execute
	s, err := gochroma.SerializeColorOutPoint(cop)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	dcop, err := gochroma.DeserializeColorOutPoint(s)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(dcop.Id, wantId) != 0 {
		t.Fatalf("didn't get what we expected: want %v, got %v", wantId, dcop.Id)
	}
	if bytes.Compare(dcop.Color, wantId) != 0 {
		t.Fatalf("didn't get what we expected: want %v, got %v", wantId, dcop.Color)
	}
	if bytes.Compare(dcop.Tx, txHash) != 0 {
		t.Fatalf("didn't get what we expected: want %v, got %v", txHash, dcop.Tx)
	}
	if bytes.Compare(dcop.SpendingTx, txHash) != 0 {
		t.Fatalf("didn't get what we expected: want %v, got %v", txHash, dcop.SpendingTx)
	}
	if bytes.Compare(dcop.PkScript, txHash) != 0 {
		t.Fatalf("didn't get what we expected: want %v, got %v", txHash, dcop.PkScript)
	}
	if dcop.Index != cop.Index {
		t.Fatalf("didn't get what we expected: want %v, got %v", cop.Index, dcop.Index)
	}
	if dcop.Value != cop.Value {
		t.Fatalf("didn't get what we expected: want %v, got %v", cop.Value, dcop.Value)
	}
	if dcop.ColorValue != cop.ColorValue {
		t.Fatalf("didn't get what we expected: want %v, got %v", cop.ColorValue, dcop.ColorValue)
	}
	if dcop.SpendingIndex != cop.SpendingIndex {
		t.Fatalf("didn't get what we expected: want %v, got %v", cop.SpendingIndex, dcop.SpendingIndex)
	}
}
