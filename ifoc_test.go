package gochroma_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/conformal/btcutil"
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

func TestColorInsValid(t *testing.T) {
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}

	genesisBytes := "01000000011a932892802d8e1657bdc84feb3663a38ea64c33b0f5436606309d6f610a01fd000000006b483045022100d1fdca93b2074caf8fe329babe0472d381721384f183566cdf7ea34e8522df3402203e0176301ef6a192bccf94ae1c5ed50df67e1ff2227fb35b6da6adafbcc1321901210277d7813a44ee7325b9cdffd22e9b0f44ad3b5b0433cc69853c21cc7e6ebeb503ffffffff0210270000000000001976a914143caef14f63625b633b77dedac55f9deaedae6088acf0908800000000001976a9147d83495938585f3f9e01cfb2137f94b0f0f2ce2588ac00000000"
	txBytesList := []string{genesisBytes, genesisBytes, genesisBytes}
	rawTx := make([][]byte, len(txBytesList))
	for i, str := range txBytesList {
		bytes, err := hex.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to convert string to bytes")
		}
		rawTx[i] = bytes
	}
	txBlockHashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	txBlockList := []string{txBlockHashStr, txBlockHashStr}
	txBlockHash := make([][]byte, len(txBlockList))
	for i, str := range txBlockList {
		bytes, err := hex.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to convert string to bytes")
		}
		txBlockHash[i] = bytes
	}
	blockBytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	blockBytesList := []string{blockBytesStr, blockBytesStr}
	block := make([][]byte, len(blockBytesList))
	for i, str := range blockBytesList {
		bytes, err := hex.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to convert string to bytes")
		}
		block[i] = bytes
	}
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: txBlockHash,
		block:       block,
		rawTx:       rawTx,
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(rawTx[0])
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	genesis := &tx.MsgTx().TxIn[0].PreviousOutPoint
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)
	colorIn := gochroma.ColorIn{outPoint, gochroma.ColorValue(1)}
	colorIns := []*gochroma.ColorIn{&colorIn}
	verify, err := ifoc.ColorInsValid(b, genesis, colorIns)
	if err != nil {
		t.Fatalf("failed with %v", err)
	}
	if verify == false {
		t.Fatalf("failed to verify: %v", colorIns)
	}
}

func TestOutPointToColorIn(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}

	genesisBytes := "01000000011a932892802d8e1657bdc84feb3663a38ea64c33b0f5436606309d6f610a01fd000000006b483045022100d1fdca93b2074caf8fe329babe0472d381721384f183566cdf7ea34e8522df3402203e0176301ef6a192bccf94ae1c5ed50df67e1ff2227fb35b6da6adafbcc1321901210277d7813a44ee7325b9cdffd22e9b0f44ad3b5b0433cc69853c21cc7e6ebeb503ffffffff0210270000000000001976a914143caef14f63625b633b77dedac55f9deaedae6088acf0908800000000001976a9147d83495938585f3f9e01cfb2137f94b0f0f2ce2588ac00000000"
	txBytesList := []string{genesisBytes, genesisBytes, genesisBytes}
	rawTx := make([][]byte, len(txBytesList))
	for i, str := range txBytesList {
		bytes, err := hex.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to convert string to bytes")
		}
		rawTx[i] = bytes
	}
	txBlockHashStr := "00000000003583bc221e70c80ce8e3d67b49be70bb3b1fd6a191d2040babd3e8"
	txBlockList := []string{txBlockHashStr, txBlockHashStr}
	txBlockHash := make([][]byte, len(txBlockList))
	for i, str := range txBlockList {
		bytes, err := hex.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to convert string to bytes")
		}
		txBlockHash[i] = bytes
	}
	blockBytesStr := "020000009153031afe12d843b71b2a8a64ba0c516630e5fe34ee0a228d4b0400000000003f38188e708f2af4973972100e29b221c3c7c703ce12ad4c42d469aaf8267f2cc2e12e54c0ff3f1b1cc2312f0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2303057b04164b6e434d696e657242519dceb367fae996d0542ee1c200000000a0010000ffffffff0100f90295000000001976a9149e8985f82bc4e0f753d0492aa8d11cc39925774088ac00000000"
	blockBytesList := []string{blockBytesStr, blockBytesStr}
	block := make([][]byte, len(blockBytesList))
	for i, str := range blockBytesList {
		bytes, err := hex.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to convert string to bytes")
		}
		block[i] = bytes
	}
	blockReaderWriter := &TstBlockReaderWriter{
		txBlockHash: txBlockHash,
		block:       block,
		rawTx:       rawTx,
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}
	tx, err := btcutil.NewTxFromBytes(rawTx[0])
	if err != nil {
		t.Fatalf("failed to get tx %v", err)
	}
	genesis := &tx.MsgTx().TxIn[0].PreviousOutPoint
	outPoint := btcwire.NewOutPoint(tx.Sha(), 0)

	// Execute
	colorIn, err := ifoc.OutPointToColorIn(b, genesis, outPoint)
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

func TestIssuingTxError(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}
	tests := []struct {
		desc       string
		bytes      []string
		fee        int64
		colorValue gochroma.ColorValue
		colorOuts  int
		err        int
	}{
		{
			desc:       "block read error",
			bytes:      []string{},
			fee:        100,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			err:        gochroma.ErrBlockRead,
		},
		{
			desc:       "negative fee",
			bytes:      []string{"0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"},
			fee:        -1,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			err:        gochroma.ErrNegativeValue,
		},
		{
			desc:       "insufficient funds",
			bytes:      []string{"0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"},
			fee:        100000000,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  1,
			err:        gochroma.ErrInsufficientFunds,
		},
		{
			desc:       "too much color value",
			bytes:      []string{"0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"},
			fee:        100,
			colorValue: gochroma.ColorValue(2),
			colorOuts:  1,
			err:        gochroma.ErrInsufficientColorValue,
		},
		{
			desc:       "too little color value",
			bytes:      []string{"0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"},
			fee:        100,
			colorValue: gochroma.ColorValue(0),
			colorOuts:  1,
			err:        gochroma.ErrDestroyColorValue,
		},
		{
			desc:       "multiple outputs",
			bytes:      []string{"0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0200e1f505000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000"},
			fee:        100,
			colorValue: gochroma.ColorValue(1),
			colorOuts:  2,
			err:        gochroma.ErrInvalidColorValue,
		},
	}

	for _, test := range tests {
		rawTx := make([][]byte, len(test.bytes))
		for i, str := range test.bytes {
			bytes, _ := hex.DecodeString(str)
			rawTx[i] = bytes
		}
		blockReaderWriter := &TstBlockReaderWriter{
			rawTx: rawTx,
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
		_, err = ifoc.IssuingTx(b, inputs, outputs, change, test.fee)

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
	tx, err := ifoc.TransferringTx(b, inputs, outputs, change, fee, false)
	if err != nil {
		t.Fatalf("error transferring  tx: %v", err)
	}

	// Verify
	if len(tx.TxOut) != 2 {
		t.Fatalf("wrong number of tx outs: got %d want %d", len(tx.TxOut))
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

func TestTransferringTxError(t *testing.T) {
	// Setup
	ifoc, err := gochroma.GetColorKernel(key)
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
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
			err:      gochroma.ErrInvalidColorValue,
		},
		{
			desc:     "bad output",
			inValue:  gochroma.ColorValue(1),
			outValue: gochroma.ColorValue(2),
			fee:      100,
			err:      gochroma.ErrInvalidColorValue,
		},
		{
			desc:     "insufficient funds",
			inValue:  gochroma.ColorValue(1),
			outValue: gochroma.ColorValue(2),
			fee:      100000000,
			err:      gochroma.ErrInsufficientFunds,
		},
	}

	for _, test := range tests {
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
		_, err = ifoc.TransferringTx(b, inputs, outputs, change, test.fee, false)

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
		outputs, err := ifoc.CalculateOutColorValues(genesis, msgTx, test.inputs)
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
		_, err = ifoc.CalculateOutColorValues(genesis, msgTx, test.inputs)

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
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	inputs, err := ifoc.FindAffectingInputs(b, genesis, msgTx, outputs)
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
	blockReaderWriter := &TstBlockReaderWriter{}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	// Execute
	inputs, err := ifoc.FindAffectingInputs(b, genesis, msgTx, nil)
	if err != nil {
		t.Fatalf("err on calculating out color values: %v", err)
	}

	// Verify
	if len(inputs) != 0 {
		t.Fatalf("wrong number of inputs: got %v, want 0", len(inputs))
	}
}

func TestAffecting(t *testing.T) {
	tests := []struct {
		desc      string
		bytes     string
		numInputs int
	}{
		{
			desc:      "normal",
			bytes:     "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0210270000000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000",
			numInputs: 1,
		}, {
			desc:      "wrong amount",
			bytes:     "0100000001aa570d9d285fe85030361b9704068b80bea89e49ad26079c2ecca8a555f8bbb8010000006c493046022100b09a37ead2637d8ffdbe2fb896a74a1c9e2f01ce306b24def2688cb7810ae609022100c019910aaf0a3317d4555441580bc5a5de6f7851d86e81aa854fef38debfefbc0121037843af5cf98718f57d6887f01d7b30bd0c6ed915eb6648ee30889861bd3a7feaffffffff0211270000000000001976a9149bbd3b6b3da61901454a9e3c0a22ac6c626cc0fa88ac32f8196f000000001976a9144d273d3a2ce1824d1c6db0764eebb03f368fd9af88ac00000000",
			numInputs: 0,
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
		ifocKernel, err := gochroma.GetColorKernel(key)
		if err != nil {
			t.Errorf("%v: error getting ifoc kernel: %v", test.desc, err)
			continue
		}
		ifoc := ifocKernel.(*gochroma.IFOC)
		txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
		msgTx.AddTxOut(txOut)
		rand.Read(hashBytes)
		genesisShaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		genesis := btcwire.NewOutPoint(genesisShaHash, 0)
		outputs := []int{0}
		bytesWant, _ := hex.DecodeString(test.bytes)
		blockReaderWriter := &TstBlockReaderWriter{
			rawTx: [][]byte{bytesWant},
		}
		b := &gochroma.BlockExplorer{blockReaderWriter}

		// Execute
		inputs, err := ifoc.FindAffectingInputs(b, genesis, msgTx, outputs)
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
		{
			desc:    "block read error",
			outputs: []int{0},
			err:     gochroma.ErrBlockRead,
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
		ifocKernel, err := gochroma.GetColorKernel(key)
		if err != nil {
			t.Errorf("%v: error getting ifoc kernel: %v", test.desc, err)
			continue
		}
		ifoc := ifocKernel.(*gochroma.IFOC)
		txOut := btcwire.NewTxOut(ifoc.TransferAmount, nil)
		msgTx.AddTxOut(txOut)
		rand.Read(hashBytes)
		genesisShaHash, err := btcwire.NewShaHash(hashBytes)
		if err != nil {
			t.Errorf("%v: err on shahash creation: %v", test.desc, err)
			continue
		}
		genesis := btcwire.NewOutPoint(genesisShaHash, 0)
		blockReaderWriter := &TstBlockReaderWriter{}
		b := &gochroma.BlockExplorer{blockReaderWriter}

		// Execute
		_, err = ifoc.FindAffectingInputs(b, genesis, msgTx, test.outputs)

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
