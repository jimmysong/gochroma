package gochroma_test

import (
	"crypto/rand"
	"testing"

	"github.com/monetas/btcwire"
	"github.com/monetas/gochroma"
)

func TestRegisterColorKernelError(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel("IFOC")
	if err != nil {
		t.Fatalf("failed to get ifoc kernel: %v", err)
	}

	// Execute
	err = gochroma.RegisterColorKernel(ifoc)

	// Verify
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrDuplicateKernel)
	if rerr.ErrorCode != wantErr {
		t.Fatalf("got wrong error: got %v, want %v", rerr.ErrorCode, wantErr)
	}
}

func TestGetColorKernelError(t *testing.T) {
	// Execute
	_, err := gochroma.GetColorKernel("NONSENSE")

	// Verify
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrNonExistentKernel)
	if rerr.ErrorCode != wantErr {
		t.Fatalf("got wrong error: got %v, want %v", rerr.ErrorCode, wantErr)
	}
}

func TestNewColorDefinition(t *testing.T) {
	// Setup
	kernel, err := gochroma.GetColorKernel("IFOC")
	if err != nil {
		t.Fatalf("failed to get ifoc kernel: %v", err)
	}
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(shaHash, 0)

	// Execute
	cd, err := gochroma.NewColorDefinition(kernel, genesis, int64(1))
	if err != nil {
		t.Fatalf("err on color definition creation: %v", err)
	}

	// Verify
	wantStr := "IFOC:" + shaHash.String() + ":0:1"
	if cd.String() != wantStr {
		t.Fatalf("wrong definition, got: %v, want %v", cd.String(), wantStr)
	}
}

func TestRunKernel(t *testing.T) {
	// Setup
	ifocKernel, err := gochroma.GetColorKernel("IFOC")
	if err != nil {
		t.Fatalf("failed to get ifoc kernel: %v", err)
	}
	ifoc := ifocKernel.(*gochroma.IFOC)
	cdStr := "IFOC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:1"
	cd, err := gochroma.NewColorDefinitionFromStr(cdStr)
	if err != nil {
		t.Fatalf("err on color definition creation: %v", err)
	}
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
	txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
	msgTx.AddTxOut(txOut)

	// Execute
	outputs, err := cd.RunKernel(msgTx, []gochroma.ColorValue{1})
	if err != nil {
		t.Fatalf("err on running kernel: %v", err)
	}

	// Verify
	if len(outputs) != 1 || outputs[0] != gochroma.ColorValue(1) {
		t.Fatalf("wrong output, got: %v, want %v", outputs, []gochroma.ColorValue{1})
	}
}

func TestNewColorDefinitionFromStr(t *testing.T) {
	// Setup
	cdStr := "IFOC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:1"

	// Execute
	cd, err := gochroma.NewColorDefinitionFromStr(cdStr)
	if err != nil {
		t.Fatalf("err on color definition creation: %v", err)
	}

	// Verify
	if cd.String() != cdStr {
		t.Fatalf("wrong definition, got: %v, want %v", cd.String(), cdStr)
	}
}

func TestNewColorDefinitionFromStrError(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		err   int
	}{
		{
			desc:  "too few components",
			input: "IFOC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:1",
			err:   gochroma.ErrBadColorDefinition,
		},
		{
			desc:  "non existent kernel",
			input: "NONSENSE:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:1",
			err:   gochroma.ErrNonExistentKernel,
		},
		{
			desc:  "invalid hash",
			input: "IFOC:xxx:0:1",
			err:   gochroma.ErrInvalidTx,
		},
		{
			desc:  "invalid index",
			input: "IFOC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:a:1",
			err:   gochroma.ErrInvalidTx,
		},
		{
			desc:  "invalid height",
			input: "IFOC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:a",
			err:   gochroma.ErrInvalidTx,
		},
		{
			desc:  "negative height",
			input: "IFOC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:-1",
			err:   gochroma.ErrInvalidTx,
		},
	}

	for _, test := range tests {
		// Execute
		_, err := gochroma.NewColorDefinitionFromStr(test.input)

		// Verify
		if err == nil {
			t.Errorf("%v: expected error got nil", test.desc)
			continue
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
		}
	}
}
