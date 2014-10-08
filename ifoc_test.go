package gochroma_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

var key = "IFOC"

func TestCode(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}

	// Execute
	str := ifoc.Code()

	// Verify
	if str != key {
		t.Fatalf("wrong KernelCode, got: %v, want %v", str, key)
	}
}

func TestIssuingTx(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("failed to convert hash %x: %v", hashBytes, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)
	inputs := []*btcwire.OutPoint{outPoint}
	initial := make([]byte, 32)
	rand.Read(initial)
	amount := gochroma.ColorValue(1)
	outputs := []*gochroma.ColorOut{&gochroma.ColorOut{initial, amount}}
	change := make([]byte, 32)
	rand.Read(change)
	fee := int64(100)

	// Execute
	tx, err := ifoc.IssuingTx(b, inputs, outputs, change, fee)
	if err != nil {
		t.Fatalf("error issuing tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut), 2)
	}
	output1 := tx.TxOut[0].Value
	output2 := tx.TxOut[1].Value
	ifoc2 := ifoc.(*gochroma.IFOC)

	wantValue := ifoc2.TransferAmount
	if output1 != wantValue {
		t.Fatalf("wrong amount in first output: got %d, want %d",
			output1, wantValue)
	}
	wantValue = int64(100000000) - ifoc2.TransferAmount - fee
	if output2 != wantValue {
		t.Fatalf("wrong amount in second output: got %d, want %d",
			output2, wantValue)
	}
}

func TestTransferringTx(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	bytesStr := "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"
	bytesWant, _ := hex.DecodeString(bytesStr)
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{bytesWant},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("failed to convert hash %x: %v", hashBytes, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)
	inputs := []*gochroma.ColorIn{&gochroma.ColorIn{outPoint, gochroma.ColorValue(1)}}
	outScript := make([]byte, 32)
	rand.Read(outScript)
	outputs := []*gochroma.ColorOut{&gochroma.ColorOut{outScript, gochroma.ColorValue(1)}}
	change := make([]byte, 32)
	rand.Read(change)
	fee := int64(100)

	// Execute
	tx, err := ifoc.TransferringTx(b, inputs, outputs, change, fee, false)
	if err != nil {
		t.Fatalf("error issuing tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut), 2)
	}
	output1 := tx.TxOut[0].Value
	output2 := tx.TxOut[1].Value
	ifoc2 := ifoc.(*gochroma.IFOC)

	wantValue := ifoc2.TransferAmount
	if output1 != wantValue {
		t.Fatalf("wrong amount in first output: got %d, want %d",
			output1, wantValue)
	}
	wantValue = int64(100000000) - ifoc2.TransferAmount - fee
	if output2 != wantValue {
		t.Fatalf("wrong amount in second output: got %d, want %d",
			output2, wantValue)
	}
}

func TestCalculateGenesis(t *testing.T) {
	// Setup
	msgTx := btcwire.NewMsgTx()
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	prevOut := btcwire.NewOutPoint(shaHash, 0)
	txIn := btcwire.NewTxIn(prevOut, nil)
	msgTx.AddTxIn(txIn)
	ifocKernel, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)
	txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, err := msgTx.TxSha()
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(&genesisShaHash, 0)
	inputs := []gochroma.ColorValue{1}

	// Execute
	outputs, err := ifoc.CalculateOutColorValues(genesis, msgTx, inputs)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(outputs) != 1 {
		t.Fatalf("wrong number of outputs: got %v, want 1", len(outputs))
	}
	if outputs[0] != gochroma.ColorValue(1) {
		t.Fatalf("wrong output value: got %v, want 1", outputs[0])
	}
}

func TestCalculate(t *testing.T) {
	ifocKernel, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)

	tests := []struct {
		desc           string
		inputs         []gochroma.ColorValue
		outputs        []gochroma.ColorValue
		firstOutAmount int64
	}{
		{
			desc:           "normal transfer",
			inputs:         []gochroma.ColorValue{1},
			outputs:        []gochroma.ColorValue{1},
			firstOutAmount: ifoc.TransferAmount,
		},
		{
			desc:           "multiple transfer",
			inputs:         []gochroma.ColorValue{1, 0, 0, 0},
			outputs:        []gochroma.ColorValue{1, 0},
			firstOutAmount: ifoc.TransferAmount,
		},
		{
			desc:           "destroy transfer",
			inputs:         []gochroma.ColorValue{1},
			outputs:        []gochroma.ColorValue{0, 0},
			firstOutAmount: int64(100),
		},
		{
			desc:           "null transfer",
			inputs:         []gochroma.ColorValue{0, 0, 0},
			outputs:        []gochroma.ColorValue{0, 0},
			firstOutAmount: ifoc.TransferAmount,
		},
	}

	for _, test := range tests {
		// Setup
		msgTx := btcwire.NewMsgTx()
		for _ = range test.inputs {
			hashBytes := make([]byte, 32)
			rand.Read(hashBytes)
			shaHash, err := btcwire.NewShaHash(hashBytes)
			if err != nil {
				t.Fatalf("err on shahash creation: %v", err)
			}
			prevOut := btcwire.NewOutPoint(shaHash, 0)
			txIn := btcwire.NewTxIn(prevOut, nil)
			msgTx.AddTxIn(txIn)
		}
		txOut := btcwire.NewTxOut(test.firstOutAmount, nil)
		msgTx.AddTxOut(txOut)
		for _ = range test.outputs[1:] {
			var txOut *btcwire.TxOut
			txOut = btcwire.NewTxOut(20000, nil)
			msgTx.AddTxOut(txOut)
		}
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Fatalf("err on shahash creation: %v", err)
		}
		genesis := btcwire.NewOutPoint(shaHash, 0)

		// Execute
		outputs, err := ifoc.CalculateOutColorValues(genesis, msgTx, test.inputs)
		if err != nil {
			t.Fatalf("%v: err on calculating out color values: %v", test.desc, err)
		}

		// Verify
		if len(outputs) != len(test.outputs) {
			t.Fatalf("%v: wrong number of outputs: got %v, want %v",
				test.desc, len(outputs), len(test.outputs))
		}
		for i, output := range outputs {
			if output != test.outputs[i] {
				t.Fatalf("%v: wrong output value at %d: got %v, want %v",
					test.desc, i, output, test.outputs[i],
				)
			}
		}
	}
}

func TestCalculateError(t *testing.T) {
	ifocKernel, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)

	tests := []struct {
		desc   string
		inputs []gochroma.ColorValue
		err    int
	}{
		{
			desc:   "too much color value",
			inputs: []gochroma.ColorValue{1, 1},
			err:    gochroma.ErrTooMuchColorValue,
		},
		{
			desc:   "bad first color value",
			inputs: []gochroma.ColorValue{0, 1},
			err:    gochroma.ErrInvalidColorValue,
		},
	}

	for _, test := range tests {
		// Setup
		msgTx := btcwire.NewMsgTx()
		for _ = range test.inputs {
			hashBytes := make([]byte, 32)
			rand.Read(hashBytes)
			shaHash, err := btcwire.NewShaHash(hashBytes)
			if err != nil {
				t.Fatalf("err on shahash creation: %v", err)
			}
			prevOut := btcwire.NewOutPoint(shaHash, 0)
			txIn := btcwire.NewTxIn(prevOut, nil)
			msgTx.AddTxIn(txIn)
		}
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Fatalf("err on shahash creation: %v", err)
		}
		genesis := btcwire.NewOutPoint(shaHash, 0)

		// Execute
		_, err = ifoc.CalculateOutColorValues(genesis, msgTx, test.inputs)

		// Verify
		if err == nil {
			t.Fatalf("%v: expected error, got nil", test.desc)
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
		}
	}
}

func TestAffectingGenesis(t *testing.T) {
	// Setup
	msgTx := btcwire.NewMsgTx()
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	prevOut := btcwire.NewOutPoint(shaHash, 0)
	txIn := btcwire.NewTxIn(prevOut, nil)
	msgTx.AddTxIn(txIn)
	ifocKernel, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)
	txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, err := msgTx.TxSha()
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(&genesisShaHash, 0)
	outputs := []int{0}

	// Execute
	inputs, err := ifoc.FindAffectingInputs(genesis, msgTx, outputs)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(inputs) != 0 {
		t.Fatalf("wrong number of inputs: got %v, want 0", len(inputs))
	}
}

func TestAffectingNil(t *testing.T) {
	// Setup
	msgTx := btcwire.NewMsgTx()
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	prevOut := btcwire.NewOutPoint(shaHash, 0)
	txIn := btcwire.NewTxIn(prevOut, nil)
	msgTx.AddTxIn(txIn)
	ifocKernel, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)
	txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
	msgTx.AddTxOut(txOut)
	rand.Read(hashBytes)
	genesisShaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(genesisShaHash, 0)

	// Execute
	inputs, err := ifoc.FindAffectingInputs(genesis, msgTx, nil)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(inputs) != 0 {
		t.Fatalf("wrong number of inputs: got %v, want 0", len(inputs))
	}
}

func TestAffecting(t *testing.T) {
	// Setup
	msgTx := btcwire.NewMsgTx()
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	prevOut := btcwire.NewOutPoint(shaHash, 0)
	txIn := btcwire.NewTxIn(prevOut, nil)
	msgTx.AddTxIn(txIn)
	ifocKernel, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)
	txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
	msgTx.AddTxOut(txOut)
	rand.Read(hashBytes)
	genesisShaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(genesisShaHash, 0)
	outputs := []int{0}

	// Execute
	inputs, err := ifoc.FindAffectingInputs(genesis, msgTx, outputs)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(inputs) != 1 {
		t.Fatalf("wrong number of inputs: got %v, want 1", len(inputs))
	}

	if inputs[0] != msgTx.TxIn[0] {
		t.Fatalf("wrong input: got %v, want %v", inputs[0], msgTx.TxIn[0])
	}
}

func TestAffectingError(t *testing.T) {

	tests := []struct {
		desc    string
		outputs []int
		err     int
	}{
		{
			desc:    "too many outputs",
			outputs: []int{0, 1},
			err:     gochroma.ErrTooManyOutputs,
		}, {
			desc:    "bad output index",
			outputs: []int{1},
			err:     gochroma.ErrBadOutputIndex,
		},
	}

	for _, test := range tests {
		// Setup
		msgTx := btcwire.NewMsgTx()
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Fatalf("err on shahash creation: %v", err)
		}
		prevOut := btcwire.NewOutPoint(shaHash, 0)
		txIn := btcwire.NewTxIn(prevOut, nil)
		msgTx.AddTxIn(txIn)
		ifocKernel, err := gochroma.GetColorKernel(key)
		if err != nil {
			t.Fatalf("error getting ifoc kernel: %v", err)
		}
		ifoc := ifocKernel.(*gochroma.IFOC)
		txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
		msgTx.AddTxOut(txOut)
		rand.Read(hashBytes)
		genesisShaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Fatalf("err on shahash creation: %v", err)
		}
		genesis := btcwire.NewOutPoint(genesisShaHash, 0)

		// Execute
		_, err = ifoc.FindAffectingInputs(genesis, msgTx, test.outputs)

		// Verify
		if err == nil {
			t.Fatalf("%v: expected error got nil", test.desc)
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
		}
	}
}
