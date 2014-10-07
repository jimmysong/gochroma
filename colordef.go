package gochroma

import (
	"fmt"

	"github.com/conformal/btcwire"
)

type ColorId uint64

type ColorValue uint64

type ColorKernel interface {
	KernelCode() string
	CalculateOutColorValues(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error)
	AffectingInputs(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, outputs []int) ([]*btcwire.TxIn, error)
}

type IFOC struct {
	TransferAmount int64
}

func (k *IFOC) KernelCode() string {
	return "IFOC"
}

func (k *IFOC) CalculateOutColorValues(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error) {

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

func (k *IFOC) AffectingInputs(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, outputs []int) ([]*btcwire.TxIn, error) {
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

var IFOCKernel = &IFOC{
	TransferAmount: 10000,
}

type ColorDefinition struct {
	ColorKernel
	Id      ColorId
	Genesis *btcwire.OutPoint
	Height  int64
}

func (c *ColorDefinition) String() string {
	return fmt.Sprintf("%v:%v:%d:%d", c.KernelCode(), c.Genesis.Hash, c.Genesis.Index, c.Height)
}

func (c *ColorDefinition) RunKernel(tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error) {
	return c.CalculateOutColorValues(c.Genesis, tx, inputs)
}
