package gochroma_test

import (
	"crypto/rand"
	"testing"

	"github.com/monetas/btcutil"
	"github.com/monetas/btcwire"
	"github.com/monetas/gochroma"
)

var (
	SPOBCKey = "SPOBC"
)

func TestSPOBCCode(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}

	// Execute
	str := spobc.Code()

	// Verify
	if str != SPOBCKey {
		t.Fatalf("wrong KernelCode, got: %v, want %v", str, SPOBCKey)
	}
}

func TestSPOBCIssuingSatoshiNeeded(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}

	// Execute
	amount := spobc.IssuingSatoshiNeeded(gochroma.ColorValue(1))

	// Verify
	spobc2 := spobc.(*gochroma.SPOBC)
	if amount != spobc2.MinimumSatoshi {
		t.Fatalf("wrong KernelCode, got: %v, want %v",
			amount, spobc2.MinimumSatoshi)
	}
}

func TestSPOBCIssuingTx(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx:       [][]byte{normalTx},
		txOutSpents: []bool{false},
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
	tx, err := spobc.IssuingTx(b, inputs, outputs, change, fee)
	if err != nil {
		t.Fatalf("error issuing tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut), 2)
	}
	output1 := tx.TxOut[0].Value
	output2 := tx.TxOut[1].Value
	spobc2 := spobc.(*gochroma.SPOBC)

	wantValue := spobc2.MinimumSatoshi
	if output1 != wantValue {
		t.Fatalf("wrong amount in first output: got %d, want %d",
			output1, wantValue)
	}
	wantValue = int64(100000000) - spobc2.MinimumSatoshi - fee
	if output2 != wantValue {
		t.Fatalf("wrong amount in second output: got %d, want %d",
			output2, wantValue)
	}

	gotMarker := tx.TxIn[0].Sequence
	if gotMarker != gochroma.SPOBCSequenceMarker {
		t.Fatalf("wrong marker in tx: got %d, want %d",
			gotMarker, gochroma.SPOBCSequenceMarker)
	}

}

func TestSPOBCIssuingTxError(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	tests := []struct {
		desc       string
		bytes      [][]byte
		fee        int64
		colorValue gochroma.ColorValue
		colorOuts  int
		spents     []bool
		err        int
	}{
		{
			desc:       "block read error",
			bytes:      [][]byte{},
			fee:        100,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrBlockRead,
		},
		{
			desc:       "negative fee",
			bytes:      [][]byte{normalTx},
			fee:        -1,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrNegativeValue,
		},
		{
			desc:       "insufficient funds",
			bytes:      [][]byte{normalTx},
			fee:        100000000,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrInsufficientFunds,
		},
		{
			desc:       "too much color value",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(2),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrInsufficientColorValue,
		},
		{
			desc:       "too little color value",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(0),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrInsufficientColorValue,
		},
		{
			desc:       "multiple outputs",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  2,
			spents:     []bool{false},
			err:        gochroma.ErrInvalidColorValue,
		},
		{
			desc:       "spent already",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			spents:     []bool{true},
			err:        gochroma.ErrOutPointSpent,
		},
		{
			desc:       "error on spent retrieval",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			spents:     nil,
			err:        gochroma.ErrBlockRead,
		},
	}

	for _, test := range tests {
		rawTx := test.bytes
		blockReaderWriter := &TstBlockReaderWriter{
			rawTx:       rawTx,
			txOutSpents: test.spents,
		}
		b := &gochroma.BlockExplorer{blockReaderWriter}
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: failed to convert hash %x: %v",
				test.desc, hashBytes, err)
			continue
		}
		outPoint := btcwire.NewOutPoint(shaHash, 0)
		inputs := []*btcwire.OutPoint{outPoint}
		initial := make([]byte, 32)
		rand.Read(initial)
		outputs := make([]*gochroma.ColorOut, test.colorOuts)
		for i := 0; i < test.colorOuts; i++ {
			outputs[i] = &gochroma.ColorOut{initial, test.colorValue}
		}
		change := make([]byte, 32)
		rand.Read(change)

		// Execute
		_, err = spobc.IssuingTx(b, inputs, outputs, change, test.fee)

		// Verify
		if err == nil {
			t.Errorf("%v: got nil where we expected err", test.desc)
			continue
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
			continue
		}
	}
}

func TestSPOBCTransferringTx(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx:       [][]byte{normalTx},
		txOutSpents: []bool{false},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	shaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("failed to convert hash %x: %v", hashBytes, err)
	}
	outPoint := btcwire.NewOutPoint(shaHash, 0)
	inputs := []*gochroma.ColorIn{
		&gochroma.ColorIn{outPoint, gochroma.ColorValue(1)}}
	outScript := make([]byte, 32)
	rand.Read(outScript)
	outputs := []*gochroma.ColorOut{
		&gochroma.ColorOut{outScript, gochroma.ColorValue(1)}}
	change := make([]byte, 32)
	rand.Read(change)
	fee := int64(100)

	// Execute
	tx, err := spobc.TransferringTx(b, inputs, outputs, change, fee, false)
	if err != nil {
		t.Fatalf("error transferring  tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut))
	}
	output1 := tx.TxOut[0].Value
	output2 := tx.TxOut[1].Value
	spobc2 := spobc.(*gochroma.SPOBC)

	wantValue := spobc2.MinimumSatoshi
	if output1 != wantValue {
		t.Fatalf("wrong amount in first output: got %d, want %d",
			output1, wantValue)
	}
	wantValue = int64(100000000) - spobc2.MinimumSatoshi - fee
	if output2 != wantValue {
		t.Fatalf("wrong amount in second output: got %d, want %d",
			output2, wantValue)
	}
}

func TestSPOBCTransferringTxError(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	tests := []struct {
		desc     string
		inValue  gochroma.ColorValue
		outValue gochroma.ColorValue
		fee      int64
		err      int
	}{
		{
			desc:     "bad input",
			inValue:  gochroma.ColorValue(2),
			outValue: gochroma.ColorValue(2),
			fee:      100,
			err:      gochroma.ErrTooMuchColorValue,
		},
		{
			desc:     "bad output",
			inValue:  gochroma.ColorValue(1),
			outValue: gochroma.ColorValue(2),
			fee:      100,
			err:      gochroma.ErrInsufficientColorValue,
		},
		{
			desc:     "insufficient funds",
			inValue:  gochroma.ColorValue(1),
			outValue: gochroma.ColorValue(2),
			fee:      100000000,
			err:      gochroma.ErrInsufficientColorValue,
		},
		{
			desc:     "destroy color",
			inValue:  gochroma.ColorValue(1),
			outValue: gochroma.ColorValue(0),
			fee:      100,
			err:      gochroma.ErrDestroyColorValue,
		},
		{
			desc:     "no inputs",
			inValue:  gochroma.ColorValue(0),
			outValue: gochroma.ColorValue(0),
			fee:      100,
			err:      gochroma.ErrInsufficientColorValue,
		},
		{
			desc:     "negative fee",
			inValue:  gochroma.ColorValue(1),
			outValue: gochroma.ColorValue(1),
			fee:      -100,
			err:      gochroma.ErrNegativeValue,
		},
	}

	for _, test := range tests {
		blockReaderWriter := &TstBlockReaderWriter{
			rawTx:       [][]byte{normalTx},
			txOutSpents: []bool{false},
		}
		b := &gochroma.BlockExplorer{blockReaderWriter}
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Fatalf("%v: failed to convert hash %x: %v", test.desc, hashBytes, err)
		}
		outPoint := btcwire.NewOutPoint(shaHash, 0)
		inputs := []*gochroma.ColorIn{
			&gochroma.ColorIn{outPoint, test.inValue}}
		outScript := make([]byte, 32)
		rand.Read(outScript)
		outputs := []*gochroma.ColorOut{
			&gochroma.ColorOut{outScript, test.outValue}}
		change := make([]byte, 32)
		rand.Read(change)

		// Execute
		_, err = spobc.TransferringTx(b, inputs, outputs, change, test.fee, false)

		// Verify
		if err == nil {
			t.Errorf("%v: got nil where we expected err", test.desc)
			continue
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
			continue
		}
	}
}

func TestSPOBCCalculateGenesis(t *testing.T) {
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
	spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)
	txOut := btcwire.NewTxOut(spobc.MinimumSatoshi, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, err := msgTx.TxSha()
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(&genesisShaHash, 0)
	inputs := []gochroma.ColorValue{1}

	// Execute
	outputs, err := spobc.CalculateOutColorValues(genesis, msgTx, inputs)
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

func TestSPOBCCalculate(t *testing.T) {
	spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)

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
			firstOutAmount: spobc.MinimumSatoshi,
		},
		{
			desc:           "multiple transfer",
			inputs:         []gochroma.ColorValue{1, 0, 0, 0},
			outputs:        []gochroma.ColorValue{1, 0},
			firstOutAmount: spobc.MinimumSatoshi,
		},
		{
			desc:           "destroy transfer",
			inputs:         []gochroma.ColorValue{0, 1},
			outputs:        []gochroma.ColorValue{0},
			firstOutAmount: int64(100),
		},
		{
			desc:           "null transfer",
			inputs:         []gochroma.ColorValue{0, 0, 0},
			outputs:        []gochroma.ColorValue{0, 0},
			firstOutAmount: spobc.MinimumSatoshi,
		},
	}

OUTER:
	for _, test := range tests {
		// Setup
		msgTx := btcwire.NewMsgTx()
		for _ = range test.inputs {
			hashBytes := make([]byte, 32)
			rand.Read(hashBytes)
			shaHash, err := btcwire.NewShaHash(hashBytes)
			if err != nil {
				t.Errorf("%v: err on shahash creation: %v", test.desc, err)
				continue OUTER
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
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		genesis := btcwire.NewOutPoint(shaHash, 0)

		// Execute
		outputs, err := spobc.CalculateOutColorValues(genesis, msgTx, test.inputs)
		if err != nil {
			t.Errorf("%v: err on calculating out color values: %v",
				test.desc, err)
			continue
		}

		// Verify
		if len(outputs) != len(test.outputs) {
			t.Errorf("%v: wrong number of outputs: got %v, want %v",
				test.desc, len(outputs), len(test.outputs))
			continue
		}
		for i, output := range outputs {
			if output != test.outputs[i] {
				t.Errorf("%v: wrong output value at %d: got %v, want %v",
					test.desc, i, output, test.outputs[i],
				)
				continue OUTER
			}
		}
	}
}

func TestSPOBCCalculateError(t *testing.T) {
	spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)

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
			desc:   "too much color value",
			inputs: []gochroma.ColorValue{2, 0},
			err:    gochroma.ErrTooMuchColorValue,
		},
	}

OUTER:
	for _, test := range tests {
		// Setup
		msgTx := btcwire.NewMsgTx()
		for _ = range test.inputs {
			hashBytes := make([]byte, 32)
			rand.Read(hashBytes)
			shaHash, err := btcwire.NewShaHash(hashBytes)
			if err != nil {
				t.Errorf("%v: err on shahash creation: %v", test.desc, err)
				continue OUTER
			}
			prevOut := btcwire.NewOutPoint(shaHash, 0)
			txIn := btcwire.NewTxIn(prevOut, nil)
			msgTx.AddTxIn(txIn)
		}
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		genesis := btcwire.NewOutPoint(shaHash, 0)

		// Execute
		_, err = spobc.CalculateOutColorValues(genesis, msgTx, test.inputs)

		// Verify
		if err == nil {
			t.Errorf("%v: expected error, got nil", test.desc)
			continue
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(test.err)
		if rerr.ErrorCode != wantErr {
			t.Errorf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
			continue
		}
	}
}

func TestSPOBCAffectingGenesis(t *testing.T) {
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
	spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)
	txOut := btcwire.NewTxOut(spobc.MinimumSatoshi, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, err := msgTx.TxSha()
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(&genesisShaHash, 0)
	outputs := []int{0}

	// Execute
	inputs, err := spobc.FindAffectingInputs(nil, genesis, msgTx, outputs)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(inputs) != 0 {
		t.Fatalf("wrong number of inputs: got %v, want 0", len(inputs))
	}
}

func TestSPOBCAffectingNil(t *testing.T) {
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
	spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)
	txOut := btcwire.NewTxOut(spobc.MinimumSatoshi, nil)
	msgTx.AddTxOut(txOut)
	rand.Read(hashBytes)
	genesisShaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(genesisShaHash, 0)

	// Execute
	inputs, err := spobc.FindAffectingInputs(nil, genesis, msgTx, nil)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(inputs) != 0 {
		t.Fatalf("wrong number of inputs: got %v, want 0", len(inputs))
	}
}

func TestSPOBCAffecting(t *testing.T) {

	spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	spobc := spobcKernel.(*gochroma.SPOBC)
	txOut := btcwire.NewTxOut(spobc.MinimumSatoshi, nil)

	tests := []struct {
		desc          string
		txOuts        []*btcwire.TxOut
		outputIndexes []int
		numInputs     int
	}{
		{
			desc:          "normal",
			outputIndexes: []int{0},
			txOuts:        []*btcwire.TxOut{txOut},
			numInputs:     1,
		},
		{
			desc:          "nil",
			outputIndexes: []int{1},
			txOuts:        []*btcwire.TxOut{txOut, txOut},
			numInputs:     0,
		},
	}

OUTER:
	for _, test := range tests {
		// Setup
		msgTx := btcwire.NewMsgTx()
		hashBytes := make([]byte, 32)
		rand.Read(hashBytes)
		shaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		prevOut := btcwire.NewOutPoint(shaHash, 0)
		txIn := btcwire.NewTxIn(prevOut, nil)
		msgTx.AddTxIn(txIn)
		for _, txOut := range test.txOuts {
			msgTx.AddTxOut(txOut)
		}
		rand.Read(hashBytes)
		genesisShaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		genesis := btcwire.NewOutPoint(genesisShaHash, 0)

		// Execute
		inputs, err := spobc.FindAffectingInputs(nil, genesis, msgTx, test.outputIndexes)
		if err != nil {
			t.Errorf("%v: err on calculating out color values: %v", test.desc, err)
			continue
		}

		// Verify
		if len(inputs) != test.numInputs {
			t.Errorf("%v: wrong number of inputs: got %v, want %d",
				test.desc, len(inputs), test.numInputs)
			continue
		}
		for i, input := range inputs {
			prev := &msgTx.TxIn[i].PreviousOutPoint
			if !input.Hash.IsEqual(&prev.Hash) || input.Index != prev.Index {
				t.Errorf("%v: wrong input: got %v, want %v",
					test.desc, input, prev)
				continue OUTER
			}
		}
	}
}

func TestSPOBCAffectingError(t *testing.T) {

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
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		prevOut := btcwire.NewOutPoint(shaHash, 0)
		txIn := btcwire.NewTxIn(prevOut, nil)
		msgTx.AddTxIn(txIn)
		spobcKernel, err := gochroma.GetColorKernel(SPOBCKey)
		if err != nil {
			t.Errorf("%v: error getting spobc kernel: %v", test.desc, err)
			continue
		}
		spobc := spobcKernel.(*gochroma.SPOBC)
		txOut := btcwire.NewTxOut(spobc.MinimumSatoshi, nil)
		msgTx.AddTxOut(txOut)
		rand.Read(hashBytes)
		genesisShaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		genesis := btcwire.NewOutPoint(genesisShaHash, 0)

		// Execute
		_, err = spobc.FindAffectingInputs(nil, genesis, msgTx, test.outputs)

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
			continue
		}
	}
}

func TestSPOBCOutPointToColorIn(t *testing.T) {
	// Setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash, blockHash},
		block:       [][]byte{rawBlock, rawBlock},
		rawTx:       [][]byte{genesisTx, genesisTx},
		txOutSpents: []bool{false},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	genesis := &tx.MsgTx().TxIn[0].PreviousOutPoint
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)

	// Execute
	colorIn, err := spobc.OutPointToColorIn(b, genesis, outPoint)
	if err != nil {
		t.Fatalf("failed with %v", err)
	}

	// Validate
	cvGot := colorIn.ColorValue
	cvWant := gochroma.ColorValue(1)
	if cvGot != cvWant {
		t.Fatalf("results differ got %v, want %v", cvGot, cvWant)
	}
}

func TestSPOBCOutPointToColorInError(t *testing.T) {

	// setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	tx, err := btcutil.NewTxFromBytes(normalTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	genesis := &tx.MsgTx().TxIn[0].PreviousOutPoint
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)

	tests := []struct {
		desc        string
		blockReader TstBlockReaderWriter
	}{
		{
			desc:        "outpointspent",
			blockReader: TstBlockReaderWriter{},
		},
		{
			desc: "outpointvalue",
			blockReader: TstBlockReaderWriter{
				txOutSpents: []bool{false},
			},
		},
		{
			desc: "outpointheight",
			blockReader: TstBlockReaderWriter{
				rawTx:       [][]byte{normalTx},
				txOutSpents: []bool{false},
			},
		},
		{
			desc: "outpointheight 2",
			blockReader: TstBlockReaderWriter{
				txBlockHash: [][]byte{blockHash},
				block:       [][]byte{rawBlock},
				rawTx:       [][]byte{normalTx},
				txOutSpents: []bool{false},
			},
		},
		{
			desc: "outpointtx",
			blockReader: TstBlockReaderWriter{
				txBlockHash: [][]byte{blockHash, blockHash},
				block:       [][]byte{rawBlock, rawBlock},
				rawTx:       [][]byte{normalTx},
				txOutSpents: []bool{false},
			},
		},
	}

	for _, test := range tests {
		// execute
		_, err = spobc.OutPointToColorIn(&gochroma.BlockExplorer{&test.blockReader}, genesis, outPoint)

		// validate
		if err == nil {
			t.Fatalf("%v: expected error, got nil", test.desc)
		}
		rerr := err.(gochroma.ChromaError)
		wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
		if rerr.ErrorCode != wantErr {
			t.Fatalf("%v: wrong error passed back: got %v, want %v",
				test.desc, rerr.ErrorCode, wantErr)
		}
	}
}

func TestSPOBCColorInsValid(t *testing.T) {
	// setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)

	tests := []struct {
		colorValue  gochroma.ColorValue
		returnValue bool
	}{
		{
			colorValue:  1,
			returnValue: true,
		},
		{
			colorValue:  100,
			returnValue: false,
		},
	}

	for _, test := range tests {
		blockReaderWriter := &TstBlockReaderWriter{
			txBlockHash: [][]byte{blockHash, blockHash},
			block:       [][]byte{rawBlock, rawBlock},
			rawTx:       [][]byte{normalTx, normalTx, genesisTx},
			txOutSpents: []bool{false},
		}
		b := &gochroma.BlockExplorer{blockReaderWriter}
		colorIn := gochroma.ColorIn{outPoint, test.colorValue}
		colorIns := []*gochroma.ColorIn{&colorIn}

		// execute
		verify, err := spobc.ColorInsValid(b, outPoint, colorIns)

		// validate
		if err != nil {
			t.Fatalf("failed with %v", err)
		}
		if verify != test.returnValue {
			t.Fatalf("unexpected result: %v %v", test.colorValue, test.returnValue)
		}
	}
}

func TestSPOBCColorInsValidError(t *testing.T) {
	// setup
	spobc, err := gochroma.GetColorKernel(SPOBCKey)
	if err != nil {
		t.Fatalf("error getting spobc kernel: %v", err)
	}
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)
	colorIn := gochroma.ColorIn{outPoint, gochroma.ColorValue(0)}
	colorIns := []*gochroma.ColorIn{&colorIn}

	// execute
	_, err = spobc.ColorInsValid(b, outPoint, colorIns)

	// validate
	if err == nil {
		t.Fatalf("expected error got nil")
	}
	rerr := err.(gochroma.ChromaError)
	wantErr := gochroma.ErrorCode(gochroma.ErrBlockRead)
	if rerr.ErrorCode != wantErr {
		t.Fatalf("wrong error passed back: got %v, want %v",
			rerr.ErrorCode, wantErr)
	}
}
