package gochroma

import (
	"fmt"

	"github.com/conformal/btcwire"
)

func init() {
	RegisterColorKernel(&IFOC{TransferAmount: 10000})
}

type IFOC struct {
	TransferAmount int64
}

func (k IFOC) Code() string {
	return "IFOC"
}

func (k IFOC) getChange(b *BlockExplorer, inputs []*btcwire.OutPoint, fee int64) (*int64, error) {
	sum := int64(0)
	for _, input := range inputs {
		value, err := b.OutPointValue(input)
		if err != nil {
			str := fmt.Sprintf("input %v got errors", input)
			return nil, makeError(ErrInvalidTx, str, err)
		}
		sum += value
	}

	if fee < 0 {
		str := fmt.Sprintf("fee is negative: %d", fee)
		return nil, makeError(ErrNegativeValue, str, nil)
	}

	// add up all inputs in order and see if we have enough
	amountNeeded := fee + k.TransferAmount
	if sum < amountNeeded {
		str := fmt.Sprintf("have %d satoshi, need %d satoshi", sum,
			amountNeeded)
		return nil, makeError(ErrInsufficientFunds, str, nil)
	}
	change := sum - amountNeeded
	return &change, nil
}

func (k IFOC) checkOutputs(outputs []*ColorOut, destroy bool) error {
	if len(outputs) != 1 {
		str := fmt.Sprintf("ifoc should have exactly 1 output: %d", len(outputs))
		return makeError(ErrInvalidColorValue, str, nil)
	}

	if outputs[0].ColorValue > 1 {
		return makeError(ErrInsufficientColorValue, "ifoc only should ever have 1 color value", nil)
	}

	if !destroy && outputs[0].ColorValue < 1 {
		return makeError(ErrDestroyColorValue, "destroying color value unintentionally", nil)
	}

	return nil
}

func (k IFOC) IssuingTx(b *BlockExplorer, inputs []*btcwire.OutPoint,
	outputs []*ColorOut, changeScript []byte,
	fee int64) (*btcwire.MsgTx, error) {

	err := k.checkOutputs(outputs, false)
	if err != nil {
		return nil, err
	}

	change, err := k.getChange(b, inputs, fee)
	if err != nil {
		return nil, err
	}

	// create the transaction
	msgTx := btcwire.NewMsgTx()
	for _, input := range inputs {
		msgTx.AddTxIn(btcwire.NewTxIn(input, nil))
	}
	for _, output := range outputs {
		msgTx.AddTxOut(btcwire.NewTxOut(k.TransferAmount, output.Script))
	}
	if *change > 0 {
		msgTx.AddTxOut(btcwire.NewTxOut(*change, changeScript))
	}
	return msgTx, nil
}

func (k IFOC) TransferringTx(b *BlockExplorer, inputs []*ColorIn,
	outputs []*ColorOut, changeScript []byte,
	fee int64, destroy bool) (*btcwire.MsgTx, error) {

	k.checkOutputs(outputs, destroy)

	change, err := k.getChange(b, OutPoints(inputs), fee)
	if err != nil {
		return nil, err
	}

	// check the color value
	inSum := ColorValue(0)
	for _, input := range inputs {
		inSum += input.ColorValue
	}
	if inSum != 1 {
		return nil, makeError(ErrInvalidColorValue, "IFOC only supports exactly 1 color value", nil)
	}

	// create the transaction
	msgTx := btcwire.NewMsgTx()
	for _, input := range inputs {
		msgTx.AddTxIn(btcwire.NewTxIn(input.OutPoint, nil))
	}
	for _, output := range outputs {
		msgTx.AddTxOut(btcwire.NewTxOut(k.TransferAmount, output.Script))
	}
	if *change > 0 {
		msgTx.AddTxOut(btcwire.NewTxOut(*change, changeScript))
	}
	return msgTx, nil
}

func (k IFOC) CalculateOutColorValues(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error) {
	outputs := make([]ColorValue, len(tx.TxOut))

	// handle case where the tx is the issuing tx
	txShaHash, err := tx.TxSha()
	if err != nil {
		return nil, makeError(ErrInvalidTx, "transaction does not have a hash", err)
	}
	if genesis.Hash.String() == txShaHash.String() {
		outputs[genesis.Index] = ColorValue(1)
		return outputs, nil
	}

	// check inputs don't sum to more than 1
	sum := ColorValue(0)
	for _, value := range inputs {
		sum += value
	}
	if sum > ColorValue(1) {
		err := fmt.Sprintf("too much color value, should be 1, got %d", sum)
		return nil, makeError(ErrTooMuchColorValue, err, nil)
	} else if sum == 0 {
		return outputs, nil
	}

	// check that the first input has the 1 color value
	if inputs[0] != ColorValue(1) {
		return nil, makeError(ErrInvalidColorValue, "First Input ColorValue is not 1", nil)
	}

	// if the first tx output does not have the right transferring amount
	// we're destroying the color value
	if tx.TxOut[0].Value != k.TransferAmount {
		return outputs, nil
	}

	outputs[0] = ColorValue(1)
	return outputs, nil
}

func (k IFOC) FindAffectingInputs(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, outputs []int) ([]*btcwire.TxIn, error) {
	// handle case where the tx is the issuing tx
	txShaHash, err := tx.TxSha()
	if err != nil {
		return nil, makeError(ErrInvalidTx, "transaction does not have a hash", err)
	}
	if genesis.Hash.String() == txShaHash.String() {
		return nil, nil
	}

	// handle the case where the outputs are nil
	if outputs == nil {
		return nil, nil
	}

	// handle the case where there's more than one output
	if len(outputs) > 1 {
		return nil, makeError(ErrTooManyOutputs, "can't track back more than 1 output in IFOC", err)
	}

	// handle the case where the output is not 0
	if outputs[0] != 0 {
		return nil, makeError(ErrBadOutputIndex, "can't track back any index other than 0", err)
	}

	return []*btcwire.TxIn{tx.TxIn[0]}, nil
}
