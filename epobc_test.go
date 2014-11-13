package gochroma_test

import (
	"crypto/rand"
	"testing"

	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
)

var (
	EPOBCKey = "EPOBC"
)

func TestEPOBCCode(t *testing.T) {
	// Setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}

	// Execute
	str := epobc.Code()

	// Verify
	if str != EPOBCKey {
		t.Fatalf("wrong KernelCode, got: %v, want %v", str, EPOBCKey)
	}
}

func TestEPOBCIssuingSatoshiNeeded(t *testing.T) {

	// Setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}

	tests := []struct {
		desc string
		cv   gochroma.ColorValue
		want int64
	}{
		{
			desc: "100",
			cv:   100,
			want: 8192 + 100,
		},
		{
			desc: "1500",
			cv:   1500,
			want: 4096 + 1500,
		},
		{
			desc: "5000",
			cv:   5000,
			want: 512 + 5000,
		},
		{
			desc: "10000",
			cv:   10000,
			want: 10001,
		},
	}

	for _, test := range tests {
		// Execute
		amount := epobc.IssuingSatoshiNeeded(test.cv)

		// Verify
		if amount != test.want {
			t.Fatalf("%v: wrong amount, got: %v, want %v", test.desc, amount, test.want)
		}
	}
}

func TestEPOBCIssuingTx(t *testing.T) {
	// Setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
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
	amount := gochroma.ColorValue(10000)
	outputs := []*gochroma.ColorOut{&gochroma.ColorOut{initial, amount}}
	change := make([]byte, 32)
	rand.Read(change)
	fee := int64(100)

	// Execute
	tx, err := epobc.IssuingTx(b, inputs, outputs, change, fee)
	if err != nil {
		t.Fatalf("error issuing tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut), 2)
	}
	output1 := tx.TxOut[0].Value
	output2 := tx.TxOut[1].Value

	wantValue := int64(amount) + 1
	if output1 != wantValue {
		t.Fatalf("wrong amount in first output: got %d, want %d",
			output1, wantValue)
	}
	wantValue = int64(100000000) - wantValue - fee
	if output2 != wantValue {
		t.Fatalf("wrong amount in second output: got %d, want %d",
			output2, wantValue)
	}
	gotMarker := tx.TxIn[0].Sequence
	wantMarker := gochroma.EPOBCGenesisMarker.Combine(gochroma.NewBitList(0, 26)).Uint32()
	if gotMarker != wantMarker {
		t.Fatalf("wrong marker in tx: got %d, want %d", gotMarker, wantMarker)
	}

}

func TestEPOBCIssuingTxError(t *testing.T) {
	// Setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
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
			colorValue: gochroma.ColorValue(5000),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrBlockRead,
		},
		{
			desc:       "negative fee",
			bytes:      [][]byte{normalTx},
			fee:        -1,
			colorValue: gochroma.ColorValue(5000),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrNegativeValue,
		},
		{
			desc:       "insufficient funds",
			bytes:      [][]byte{normalTx},
			fee:        100000000,
			colorValue: gochroma.ColorValue(5000),
			colorOuts:  1,
			spents:     []bool{false},
			err:        gochroma.ErrInsufficientFunds,
		},
		{
			desc:       "multiple outputs",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(5000),
			colorOuts:  2,
			spents:     []bool{false},
			err:        gochroma.ErrInvalidColorValue,
		},
		{
			desc:       "spent already",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(5000),
			colorOuts:  1,
			spents:     []bool{true},
			err:        gochroma.ErrOutPointSpent,
		},
		{
			desc:       "error on spent retrieval",
			bytes:      [][]byte{normalTx},
			fee:        100,
			colorValue: gochroma.ColorValue(5000),
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
		_, err = epobc.IssuingTx(b, inputs, outputs, change, test.fee)

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

func TestEPOBCTransferringTx(t *testing.T) {
	// Setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
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
	cv := gochroma.ColorValue(1000)
	inputs := []*gochroma.ColorIn{
		&gochroma.ColorIn{outPoint, cv}}
	outScript := make([]byte, 32)
	rand.Read(outScript)
	outputs := []*gochroma.ColorOut{
		&gochroma.ColorOut{outScript, cv}}
	change := make([]byte, 32)
	rand.Read(change)
	fee := int64(100)

	// Execute
	tx, err := epobc.TransferringTx(b, inputs, outputs, change, fee, false)
	if err != nil {
		t.Fatalf("error transferring  tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut))
	}
	output1 := tx.TxOut[0].Value
	output2 := tx.TxOut[1].Value

	wantValue := int64(cv) + 8192
	if output1 != wantValue {
		t.Fatalf("wrong amount in first output: got %d, want %d",
			output1, wantValue)
	}
	wantValue = int64(100000000) - wantValue - fee
	if output2 != wantValue {
		t.Fatalf("wrong amount in second output: got %d, want %d",
			output2, wantValue)
	}
}

func TestEPOBCTransferringTxError(t *testing.T) {
	// Setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}
	tests := []struct {
		desc     string
		inValue  gochroma.ColorValue
		outValue gochroma.ColorValue
		fee      int64
		err      int
	}{
		{
			desc:     "bad output",
			inValue:  gochroma.ColorValue(1000),
			outValue: gochroma.ColorValue(2000),
			fee:      100,
			err:      gochroma.ErrInsufficientColorValue,
		},
		{
			desc:     "insufficient funds",
			inValue:  gochroma.ColorValue(1000),
			outValue: gochroma.ColorValue(2000),
			fee:      100000000,
			err:      gochroma.ErrInsufficientColorValue,
		},
		{
			desc:     "destroy color",
			inValue:  gochroma.ColorValue(1000),
			outValue: gochroma.ColorValue(900),
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
			desc:     "no inputs",
			inValue:  gochroma.ColorValue(100),
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
		_, err = epobc.TransferringTx(b, inputs, outputs, change, test.fee, false)

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

func TestEPOBCCalculateGenesis(t *testing.T) {
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
	epobcKernel, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}
	epobc := epobcKernel.(*gochroma.EPOBC)
	txOut := btcwire.NewTxOut(epobc.MinimumSatoshi, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, err := msgTx.TxSha()
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(&genesisShaHash, 0)
	inputs := []gochroma.ColorValue{1}

	// Execute
	outputs, err := epobc.CalculateOutColorValues(genesis, msgTx, inputs)
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

func TestEPOBCCalculate(t *testing.T) {
	epobcKernel, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}
	epobc := epobcKernel.(*gochroma.EPOBC)

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
			firstOutAmount: epobc.MinimumSatoshi,
		},
		{
			desc:           "multiple transfer",
			inputs:         []gochroma.ColorValue{1, 0, 0, 0},
			outputs:        []gochroma.ColorValue{1, 0},
			firstOutAmount: epobc.MinimumSatoshi,
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
			firstOutAmount: epobc.MinimumSatoshi,
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
		outputs, err := epobc.CalculateOutColorValues(genesis, msgTx, test.inputs)
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

func TestEPOBCCalculateError(t *testing.T) {
	epobcKernel, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}
	epobc := epobcKernel.(*gochroma.EPOBC)

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
		_, err = epobc.CalculateOutColorValues(genesis, msgTx, test.inputs)

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

func TestEPOBCAffectingIndexes(t *testing.T) {
	epobcKernel, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatal(err)
	}
	epobc := epobcKernel.(*gochroma.EPOBC)
	tests := []struct {
		desc       string
		inputs     []gochroma.ColorValue
		outputs    []int64
		padding    int64
		outIndexes []int
		inIndexes  []int
	}{
		{
			desc:       "direct",
			inputs:     []gochroma.ColorValue{1},
			outputs:    []int64{9},
			padding:    8,
			outIndexes: []int{0},
			inIndexes:  []int{0},
		},
		{
			desc:       "empty",
			inputs:     []gochroma.ColorValue{1, 2, 3},
			outputs:    []int64{1029, 1039},
			padding:    1024,
			outIndexes: nil,
			inIndexes:  nil,
		},
		{
			desc:       "join",
			inputs:     []gochroma.ColorValue{2, 3},
			outputs:    []int64{21},
			padding:    16,
			outIndexes: []int{0},
			inIndexes:  []int{0, 1},
		},
		{
			desc:       "split",
			inputs:     []gochroma.ColorValue{5},
			outputs:    []int64{34, 35},
			padding:    32,
			outIndexes: []int{0},
			inIndexes:  []int{0},
		},
		{
			desc:       "split 2",
			inputs:     []gochroma.ColorValue{5},
			outputs:    []int64{66, 67},
			padding:    64,
			outIndexes: []int{1},
			inIndexes:  []int{0},
		},
		{
			desc:       "0's",
			inputs:     []gochroma.ColorValue{0, 0, 0},
			outputs:    []int64{128, 128, 128},
			padding:    128,
			outIndexes: []int{0, 1, 2},
			inIndexes:  []int{},
		},
		{
			desc:       "null before and after",
			inputs:     []gochroma.ColorValue{0, 2, 3, 0},
			outputs:    []int64{256, 261, 299},
			padding:    256,
			outIndexes: []int{1},
			inIndexes:  []int{1, 2},
		},
		{
			desc:       "odd 1",
			inputs:     []gochroma.ColorValue{1, 2, 3},
			outputs:    []int64{517, 527},
			padding:    512,
			outIndexes: []int{0},
			inIndexes:  []int{0, 1, 2},
		},
		{
			desc:       "odd 2",
			inputs:     []gochroma.ColorValue{1, 2, 3},
			outputs:    []int64{1029, 1039},
			padding:    1024,
			outIndexes: []int{1},
			inIndexes:  []int{},
		},
	}
	for _, test := range tests {
		// setup
		var colorIns []*gochroma.ColorIn
		for _, input := range test.inputs {
			colorIns = append(colorIns, &gochroma.ColorIn{
				ColorValue: input,
			})
		}

		// execute
		inputIndexes, err := epobc.AffectingIndexes(colorIns, test.outputs, test.padding, test.outIndexes)
		if err != nil {
			t.Errorf("%v: %v", test.desc, err)
			continue
		}

		// validate
		if len(inputIndexes) != len(test.inIndexes) {
			t.Errorf("%v: input index length different: got %v want %v", test.desc, inputIndexes, test.inIndexes)
			continue
		}
		for i, inputIndex := range inputIndexes {
			if inputIndex != test.inIndexes[i] {
				t.Errorf("%v: input index different at %d: got %v want %v", test.desc, i, inputIndex, test.inIndexes[i])
				break
			}
		}
	}

}

func TestEPOBCAffectingInputs(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatal(err)
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
	txIn.Sequence = gochroma.EPOBCTransferMarker.Combine(gochroma.NewBitList(8, 26)).Uint32()
	msgTx.AddTxIn(txIn)
	txOut := btcwire.NewTxOut(100, nil)
	msgTx.AddTxOut(txOut)
	rand.Read(hashBytes)
	genesisShaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(genesisShaHash, 0)
	blockReaderWriter := &TstBlockReaderWriter{
		rawTx: [][]byte{normalTx},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// execute
	inputs, err := epobc.FindAffectingInputs(b, genesis, msgTx, []int{0})
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// validate
	if len(inputs) != 1 {
		t.Fatalf("incorrect number of inputs: want %d, got %d", 1, len(inputs))
	}
	prev := &msgTx.TxIn[0].PreviousOutPoint
	input := inputs[0]
	if !input.Hash.IsEqual(&prev.Hash) || input.Index != prev.Index {
		t.Fatalf("wrong input: got %v, want %v", input, prev)
	}
}

func TestEPOBCAffectingGenesis(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatal(err)
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
	txIn.Sequence = gochroma.EPOBCGenesisMarker.Combine(gochroma.NewBitList(8, 26)).Uint32()
	msgTx.AddTxIn(txIn)
	txOut := btcwire.NewTxOut(100, nil)
	msgTx.AddTxOut(txOut)
	genesisShaHash, err := msgTx.TxSha()
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(&genesisShaHash, 0)

	// execute
	inputs, err := epobc.FindAffectingInputs(nil, genesis, msgTx, []int{0})
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// validate
	if len(inputs) != 0 {
		t.Fatalf("incorrect number of inputs: want %d, got %d", 0, len(inputs))
	}
}

func TestEPOBCAffectingError(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatal(err)
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
	txOut := btcwire.NewTxOut(100, nil)
	msgTx.AddTxOut(txOut)
	rand.Read(hashBytes)
	genesisShaHash, err := btcwire.NewShaHash(hashBytes)
	if err != nil {
		t.Fatalf("err on shahash creation: %v", err)
	}
	genesis := btcwire.NewOutPoint(genesisShaHash, 0)
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	_, err = epobc.FindAffectingInputs(b, genesis, msgTx, []int{0})

	// Verify
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

func TestEPOBCOutPointToColorInGenesis(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}
	blockReaderWriter := &TstBlockReaderWriter{
		txOutSpents: []bool{true},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	genesis := &tx.MsgTx().TxIn[0].PreviousOutPoint
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)

	// execute
	colorIn, err := epobc.OutPointToColorIn(b, genesis, outPoint)
	if err != nil {
		t.Fatalf("failed with %v", err)
	}

	// validate
	cvGot := colorIn.ColorValue
	cvWant := gochroma.ColorValue(0)
	if cvGot != cvWant {
		t.Fatalf("results differ got %v, want %v", cvGot, cvWant)
	}
}

func TestEPOBCOutPointToColorInNormal(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
	}
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: [][]byte{blockHash, blockHash},
		block:       [][]byte{rawBlock, rawBlock},
		rawTx:       [][]byte{normalTx, normalTx, genesisTx},
		txOutSpents: []bool{false},
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(genesisTx)
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	genesis := &tx.MsgTx().TxIn[0].PreviousOutPoint
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)

	// execute
	colorIn, err := epobc.OutPointToColorIn(b, genesis, outPoint)
	if err != nil {
		t.Fatalf("failed with %v", err)
	}

	// validate
	cvGot := colorIn.ColorValue
	cvWant := gochroma.ColorValue(0)
	if cvGot != cvWant {
		t.Fatalf("results differ got %v, want %v", cvGot, cvWant)
	}
}

func TestEPOBCOutPointToColorInError(t *testing.T) {

	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
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
		{
			desc: "find affecting",
			blockReader: TstBlockReaderWriter{
				txBlockHash: [][]byte{blockHash, blockHash},
				block:       [][]byte{rawBlock, rawBlock},
				rawTx:       [][]byte{normalTx, normalTx},
				txOutSpents: []bool{false},
			},
		},
	}

	for _, test := range tests {
		// execute
		_, err = epobc.OutPointToColorIn(&gochroma.BlockExplorer{&test.blockReader}, genesis, outPoint)

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

func TestEPOBCColorInsValid(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
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
			colorValue:  0,
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
		verify, err := epobc.ColorInsValid(b, outPoint, colorIns)

		// validate
		if err != nil {
			t.Fatalf("failed with %v", err)
		}
		if verify != test.returnValue {
			t.Fatalf("unexpected result: %v %v", test.colorValue, test.returnValue)
		}
	}
}

func TestEPOBCColorInsValidError(t *testing.T) {
	// setup
	epobc, err := gochroma.GetColorKernel(EPOBCKey)
	if err != nil {
		t.Fatalf("error getting epobc kernel: %v", err)
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
	_, err = epobc.ColorInsValid(b, outPoint, colorIns)

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
