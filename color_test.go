package gochroma_test

import (
	"crypto/rand"
	"testing"

	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

func TestRegisterColorKernelError(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel("SPOBC")
	if err != nil {
		t.Fatalf("failed to get spobc kernel: %v", err)
	}

	// Execute
	err = gochroma.RegisterColorKernel(spobc)

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
	kernel, err := gochroma.GetColorKernel("SPOBC")
	if err != nil {
		t.Fatalf("failed to get spobc kernel: %v", err)
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
	wantStr := "SPOBC:" + shaHash.String() + ":0:1"
	if cd.String() != wantStr {
		t.Fatalf("wrong definition, got: %v, want %v", cd.String(), wantStr)
	}
}

func TestRunKernel(t *testing.T) {
	// Setup
	spobcKernel, err := gochroma.GetColorKernel("SPOBC")
	if err != nil {
		t.Fatalf("failed to get spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)
	cdStr := "SPOBC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:1"
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
	txOut := btcwire.NewTxOut(spobc.MinimumSatoshi, nil)
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
	cdStr := "SPOBC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:1"

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
			input: "SPOBC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:1",
			err:   gochroma.ErrBadColorDefinition,
		},
		{
			desc:  "non existent kernel",
			input: "NONSENSE:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:1",
			err:   gochroma.ErrNonExistentKernel,
		},
		{
			desc:  "invalid hash",
			input: "SPOBC:xxx:0:1",
			err:   gochroma.ErrInvalidTx,
		},
		{
			desc:  "invalid index",
			input: "SPOBC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:a:1",
			err:   gochroma.ErrInvalidTx,
		},
		{
			desc:  "invalid height",
			input: "SPOBC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:a",
			err:   gochroma.ErrInvalidTx,
		},
		{
			desc:  "negative height",
			input: "SPOBC:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:0:-1",
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
