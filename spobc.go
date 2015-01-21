package gochroma

import (
	"fmt"

	"github.com/btcsuite/btcwire"
)

var (
	// first 6 bits are 100000
	SPOBCSequenceMarker uint32 = NewBitList(1, 32).Uint32()
)

func init() {
	RegisterColorKernel(&SPOBC{MinimumSatoshi: int64(5430)})
}

type SPOBC struct {
	MinimumSatoshi int64
}

func (k SPOBC) Code() string {
	return "SPOBC"
}

func (k SPOBC) IssuingSatoshiNeeded(cv ColorValue) int64 {
	return k.MinimumSatoshi
}

func (k SPOBC) getChange(b *BlockExplorer, inputs []*btcwire.OutPoint, fee int64) (*int64, error) {
	sum := int64(0)
	for _, input := range inputs {
		// return an error if this input has been spent already
		spent, err := b.OutPointSpent(input)
		if err != nil {
			return nil, err
		}
		if *spent {
			str := fmt.Sprintf("outpoint at %v has been spent already", input)
			return nil, MakeError(ErrOutPointSpent, str, nil)
		}

		value, err := b.OutPointValue(input)
		if err != nil {
			return nil, err
		}
		sum += value
	}

	if fee < 0 {
		str := fmt.Sprintf("fee is negative: %d", fee)
		return nil, MakeError(ErrNegativeValue, str, nil)
	}

	// add up all inputs in order and see if we have enough
	amountNeeded := fee + k.MinimumSatoshi
	if sum < amountNeeded {
		str := fmt.Sprintf("have %d satoshi, need %d satoshi to issue", sum,
			amountNeeded)
		return nil, MakeError(ErrInsufficientFunds, str, nil)
	}
	change := sum - amountNeeded
	return &change, nil
}

func (k SPOBC) OutPointToColorIn(b *BlockExplorer,
	genesis, outPoint *btcwire.OutPoint) (*ColorIn, error) {

	colorIn := &ColorIn{
		OutPoint:   outPoint,
		ColorValue: ColorValue(0),
	}

	// check if this outPoint hasn't been spent already
	spent, err := b.OutPointSpent(outPoint)
	if err != nil {
		return nil, err
	}
	if *spent {
		return colorIn, nil
	}

	value, err := b.OutPointValue(outPoint)
	if err != nil {
		return nil, err
	}

	// If the outpoint is a zero-value OP_RETURN, this smart property was
	// destroyed
	if value == 0 {
		return colorIn, nil
	}
	current := outPoint
	genesisHeight, err := b.OutPointHeight(genesis)
	if err != nil {
		return nil, err
	}

	for !genesis.Hash.IsEqual(&current.Hash) {
		height, err := b.OutPointHeight(current)
		if err != nil {
			return nil, err
		}
		if height < genesisHeight {
			return colorIn, nil
		}
		tx, err := b.OutPointTx(current)
		if err != nil {
			return nil, err
		}
		inputs, err := k.FindAffectingInputs(
			b, genesis, tx.MsgTx(), []int{int(current.Index)})
		if err != nil {
			return nil, err
		}
		if inputs == nil {
			return colorIn, nil
		}
		current = inputs[0]
	}
	if current.Index == genesis.Index {
		colorIn.ColorValue = ColorValue(1)
	}
	return colorIn, nil
}

func (k SPOBC) ColorInsValid(b *BlockExplorer, genesis *btcwire.OutPoint,
	colorIns []*ColorIn) (bool, error) {
	for _, colorIn := range colorIns {
		calculated, err := k.OutPointToColorIn(b, genesis, colorIn.OutPoint)
		if err != nil {
			return false, err
		}
		if calculated.ColorValue != colorIn.ColorValue {
			return false, nil
		}
	}
	return true, nil
}

func (k SPOBC) IssuingTx(b *BlockExplorer, inputs []*btcwire.OutPoint,
	outputs []*ColorOut, changeScript []byte,
	fee int64) (*btcwire.MsgTx, error) {

	if len(outputs) != 1 {
		str := fmt.Sprintf("spobc should have exactly 1 output: %d", len(outputs))
		return nil, MakeError(ErrInvalidColorValue, str, nil)
	}

	if outputs[0].ColorValue != 1 {
		return nil, MakeError(ErrInsufficientColorValue, "spobc only should ever issue 1 color value", nil)
	}

	change, err := k.getChange(b, inputs, fee)
	if err != nil {
		return nil, err
	}

	// create the transaction
	msgTx := btcwire.NewMsgTx()
	for i, input := range inputs {
		txIn := btcwire.NewTxIn(input, nil)
		// add the special nSequence marker for the first input only
		if i == 0 {
			txIn.Sequence = SPOBCSequenceMarker
		}
		msgTx.AddTxIn(txIn)
	}
	msgTx.AddTxOut(btcwire.NewTxOut(k.MinimumSatoshi, outputs[0].Script))
	if *change > 0 {
		msgTx.AddTxOut(btcwire.NewTxOut(*change, changeScript))
	}
	return msgTx, nil
}

func (k SPOBC) TransferringTx(b *BlockExplorer, inputs []*ColorIn,
	outputs []*ColorOut, changeScript []byte,
	fee int64, destroy bool) (*btcwire.MsgTx, error) {

	sum := ColorValue(0)
	inLength := len(inputs)
	outLength := len(outputs)
	for i := 0; i < inLength && i < outLength; i++ {
		var in *ColorIn
		var out *ColorOut
		if i < inLength {
			in = inputs[i]
			sum += in.ColorValue
		}
		if i < outLength {
			out = outputs[i]
		}
		if in != nil && out != nil {
			if out.ColorValue > in.ColorValue {
				return nil, MakeError(ErrInsufficientColorValue, "you cannot create color value in a transfer", nil)
			}
		}
		if !destroy && in.ColorValue == ColorValue(1) && (out == nil || out.ColorValue == ColorValue(0)) {
			return nil, MakeError(ErrDestroyColorValue, "destroying color value unintentionally", nil)
		}
	}
	if sum > ColorValue(1) {
		return nil, MakeError(ErrTooMuchColorValue, "spobc only should ever have 1 color value", nil)
	}
	if sum == ColorValue(0) {
		return nil, MakeError(ErrInsufficientColorValue, "spobc has no color value in the inputs", nil)
	}

	change, err := k.getChange(b, OutPoints(inputs), fee)
	if err != nil {
		return nil, err
	}

	// create the transaction
	msgTx := btcwire.NewMsgTx()
	for _, input := range inputs {
		msgTx.AddTxIn(btcwire.NewTxIn(input.OutPoint, nil))
	}
	for _, output := range outputs {
		msgTx.AddTxOut(btcwire.NewTxOut(k.MinimumSatoshi, output.Script))
	}
	if *change > 0 {
		msgTx.AddTxOut(btcwire.NewTxOut(*change, changeScript))
	}
	return msgTx, nil
}

func (k SPOBC) CalculateOutColorValues(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error) {
	outputs := make([]ColorValue, len(tx.TxOut))

	// handle case where the tx is the issuing tx
	txShaHash, err := tx.TxSha()
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "transaction does not have a hash", err)
	}
	if genesis.Hash.String() == txShaHash.String() {
		outputs[genesis.Index] = ColorValue(1)
		return outputs, nil
	}

	// check inputs don't sum to more than 1
	sum := ColorValue(0)
	txOutLength := len(tx.TxOut)
	for i, value := range inputs {
		sum += value
		if value > ColorValue(1) {
			err := fmt.Sprintf("too much color value, should be at most 1, got %d", value)
			return nil, MakeError(ErrTooMuchColorValue, err, nil)
		}
		if txOutLength > i && tx.TxOut[i].Value != 0 && value == ColorValue(1) {
			outputs[i] = ColorValue(1)
		}
	}
	if sum > ColorValue(1) {
		err := fmt.Sprintf("too much color value, should be 1, got %d", sum)
		return nil, MakeError(ErrTooMuchColorValue, err, nil)
	} else if sum == 0 {
		return outputs, nil
	}

	return outputs, nil
}

func (k SPOBC) FindAffectingInputs(b *BlockExplorer, genesis *btcwire.OutPoint, tx *btcwire.MsgTx, outputIndexes []int) ([]*btcwire.OutPoint, error) {

	// note that spobc is ENTIRELY position based, so inputValues does not
	// matter at all.

	// handle case where the tx is the issuing tx
	txShaHash, err := tx.TxSha()
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "transaction does not have a hash", err)
	}
	if genesis.Hash.IsEqual(&txShaHash) {
		return nil, nil
	}

	// handle the case where the outputIndexes are nil
	if outputIndexes == nil {
		return nil, nil
	}

	// handle the case where there's more than one output
	if len(outputIndexes) > 1 {
		return nil, MakeError(ErrTooManyOutputs, "can't track back more than 1 output in SPOBC", err)
	}

	outputIndex := outputIndexes[0]

	// check that the tx has the right number of outputs
	if len(tx.TxOut) <= outputIndex {
		return nil, MakeError(ErrBadOutputIndex, "output index is corrupt", nil)
	}

	// if there aren't any inputs that correspond to the output index, send back nil
	if len(tx.TxIn) <= outputIndex {
		return nil, nil
	}

	outPoint := tx.TxIn[outputIndex].PreviousOutPoint
	return []*btcwire.OutPoint{&outPoint}, nil
}
