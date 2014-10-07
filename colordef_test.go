package gochroma_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

func TestKernelCode(t *testing.T) {
	// Setup
	ifoc := gochroma.IFOCKernel

	// Execute
	str := ifoc.KernelCode()

	// Verify
	wantStr := "IFOC"
	if str != wantStr {
		t.Fatalf("wrong KernelCode, got: %v, want %v", str, wantStr)
	}
}

func TestColorDefinitionString(t *testing.T) {
	// Setup
	hashStr := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	shaHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)
	cd := &gochroma.ColorDefinition{
		gochroma.IFOCKernel, gochroma.ColorId(1), outPoint, 1,
	}

	// Execute
	str := cd.String()

	// Verify
	wantStr := fmt.Sprintf("IFOC:%v:0:1", hashStr)
	if str != wantStr {
		t.Fatalf("wrong string, got: %v, want %v", str, wantStr)
	}
}

func TestIFOCCalcGenesis(t *testing.T) {
	// Setup
	msgTx := btcwire.NewMsgTx()
	hashStr := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	shaHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	prevOut := btcwire.NewOutPoint(shaHash, 0)
	txIn := btcwire.NewTxIn(prevOut, nil)
	msgTx.AddTxIn(txIn)
	ifoc := gochroma.IFOCKernel
	txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, _ := msgTx.TxSha()
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

func TestIFOCCalculate(t *testing.T) {
	ifoc := gochroma.IFOCKernel

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
			firstOutAmount: gochroma.IFOCKernel.TransferAmount,
		},
		{
			desc:           "multiple transfer",
			inputs:         []gochroma.ColorValue{1, 0, 0, 0},
			outputs:        []gochroma.ColorValue{1, 0},
			firstOutAmount: gochroma.IFOCKernel.TransferAmount,
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
			firstOutAmount: gochroma.IFOCKernel.TransferAmount,
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
		for i := range test.outputs {
			var txOut *btcwire.TxOut
			if i == 0 {
				txOut = btcwire.NewTxOut(test.firstOutAmount, nil)
			} else {
				txOut = btcwire.NewTxOut(20000, nil)
			}
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
