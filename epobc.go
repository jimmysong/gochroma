package gochroma

import (
	"fmt"
	"sort"

	"github.com/conformal/btcwire"
)

var (
	// first 6 bits are 101001
	EPOBCGenesisMarker = NewBitList(37, 6)
	// first 6 bits are 110011
	EPOBCTransferMarker = NewBitList(51, 6)
)

func init() {
	RegisterColorKernel(&EPOBC{MinimumSatoshi: int64(5430)})
}

type EPOBC struct {
	MinimumSatoshi int64
}

func (k EPOBC) fetchPadding(tx *btcwire.MsgTx) int64 {
	bitList := NewBitList(tx.TxIn[0].Sequence, 32)
	number := int64(0)
	digit := int64(1)
	for _, bit := range bitList[6:12] {
		if bit {
			number += digit
		}
		digit *= 2
	}
	return number
}

func (k EPOBC) Code() string {
	return "EPOBC"
}

func (k EPOBC) paddingNeeded(cv ColorValue) (BitList, int64) {
	// figure out the power of 2 that will get us the padding needed

	var padding int64
	exponent := uint32(0)
	paddingNeeded := k.MinimumSatoshi - int64(cv)
	for padding = 1; padding < paddingNeeded; padding *= 2 {
		exponent++
	}
	return NewBitList(exponent, 6), padding
}

func (k EPOBC) computeSequence(m1, m2 BitList) uint32 {
	combined := m1.Combine(m2)
	// pad with 32 - 12 = 20 more zeros
	empty := NewBitList(0, 20)
	return combined.Combine(empty).Uint32()
}

func (k EPOBC) IssuingSatoshiNeeded(cv ColorValue) int64 {
	_, padding := k.paddingNeeded(cv)
	return padding + int64(cv)
}

func (k EPOBC) getChange(b *BlockExplorer, inputs []*btcwire.OutPoint, outputs []*ColorOut, fee int64) (*int64, error) {
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

	cvSum := ColorValue(0)
	cvMin := ColorValue(k.MinimumSatoshi)
	for _, output := range outputs {
		cvSum += output.ColorValue
		if cvMin > output.ColorValue {
			cvMin = output.ColorValue
		}
	}

	// add up all inputs in order and see if we have enough
	_, padding := k.paddingNeeded(cvMin)
	amountNeeded := fee + int64(padding)*int64(len(outputs)) + int64(cvSum)

	if sum < amountNeeded {
		str := fmt.Sprintf("have %d satoshi, need %d satoshi to issue", sum,
			amountNeeded)
		return nil, MakeError(ErrInsufficientFunds, str, nil)
	}
	change := sum - amountNeeded
	return &change, nil
}

func (k EPOBC) OutPointToColorIn(b *BlockExplorer,
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

	// If the outpoint is a zero-value OP_RETURN, there's no color value
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
		// TODO: traverse all affecting inputs and get the correct color value
	}
	return colorIn, nil
}

func (k EPOBC) ColorInsValid(b *BlockExplorer, genesis *btcwire.OutPoint,
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

func (k EPOBC) IssuingTx(b *BlockExplorer, inputs []*btcwire.OutPoint,
	outputs []*ColorOut, changeScript []byte,
	fee int64) (*btcwire.MsgTx, error) {

	if len(outputs) != 1 {
		str := fmt.Sprintf("epobc should have exactly 1 output: %d", len(outputs))
		return nil, MakeError(ErrInvalidColorValue, str, nil)
	}

	change, err := k.getChange(b, inputs, outputs, fee)
	if err != nil {
		return nil, err
	}

	// create the transaction
	msgTx := btcwire.NewMsgTx()
	exponentMarker, padding := k.paddingNeeded(outputs[0].ColorValue)

	sequence := k.computeSequence(EPOBCGenesisMarker, exponentMarker)
	for i, input := range inputs {
		txIn := btcwire.NewTxIn(input, nil)
		// add the special nSequence marker for the first input only
		if i == 0 {
			txIn.Sequence = sequence
		}
		msgTx.AddTxIn(txIn)
	}
	msgTx.AddTxOut(btcwire.NewTxOut(padding+int64(outputs[0].ColorValue), outputs[0].Script))
	if *change > 0 {
		msgTx.AddTxOut(btcwire.NewTxOut(*change, changeScript))
	}
	return msgTx, nil
}

func (k EPOBC) TransferringTx(b *BlockExplorer, inputs []*ColorIn,
	outputs []*ColorOut, changeScript []byte,
	fee int64, destroy bool) (*btcwire.MsgTx, error) {

	// inputs and outputs should have non-zero color value
	inSum, outSum := ColorValue(0), ColorValue(0)
	for _, in := range inputs {
		if in.ColorValue <= 0 {
			return nil, MakeError(ErrInsufficientColorValue, "All Color Inputs should have a non-zero color value", nil)
		}
		inSum += in.ColorValue
	}
	minimum := ColorValue(1<<64 - 1)
	for _, out := range outputs {
		if out.ColorValue <= 0 {
			return nil, MakeError(ErrInsufficientColorValue, "All Color Outputs should have a non-zero color value", nil)
		}
		if minimum > out.ColorValue {
			minimum = out.ColorValue
		}
		outSum += out.ColorValue
	}

	// inputs should have more (with destroy) or equal colorvalue than outputs
	if outSum > inSum {
		return nil, MakeError(ErrInsufficientColorValue, "you cannot create color value in a transfer", nil)
	}

	if !destroy && outSum < inSum {
		return nil, MakeError(ErrDestroyColorValue, "destroying color value unintentionally", nil)
	}

	change, err := k.getChange(b, OutPoints(inputs), outputs, fee)
	if err != nil {
		return nil, err
	}

	// create the transaction
	exponentMarker, padding := k.paddingNeeded(minimum)
	sequence := k.computeSequence(EPOBCTransferMarker, exponentMarker)
	msgTx := btcwire.NewMsgTx()
	for i, input := range inputs {
		txIn := btcwire.NewTxIn(input.OutPoint, nil)
		if i == 0 {
			txIn.Sequence = sequence
		}
		msgTx.AddTxIn(txIn)
	}
	for _, output := range outputs {
		amount := int64(output.ColorValue) + padding
		txOut := btcwire.NewTxOut(amount, output.Script)
		msgTx.AddTxOut(txOut)
	}
	if *change > 0 {
		msgTx.AddTxOut(btcwire.NewTxOut(*change, changeScript))
	}
	return msgTx, nil
}

func (k EPOBC) CalculateOutColorValues(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error) {
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

func (k EPOBC) txColorIns(b *BlockExplorer, tx *btcwire.MsgTx) ([]*ColorIn, error) {
	colorIns := make([]*ColorIn, len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		prevTx, err := b.Tx(BigEndianBytes(&txIn.PreviousOutPoint.Hash))
		if err != nil {
			return nil, err
		}
		msgTx := prevTx.MsgTx()
		colorIns[i] = &ColorIn{
			OutPoint:   &txIn.PreviousOutPoint,
			ColorValue: ColorValue(msgTx.TxOut[txIn.PreviousOutPoint.Index].Value - k.fetchPadding(msgTx)),
		}
	}

	return colorIns, nil
}

// AffectingIndexes figures out which input indexes contribute to the
// output indexes as far as color values go. Exposed for testing purposes.
func (k EPOBC) AffectingIndexes(colorIns []*ColorIn, outValues []int64, padding int64, outputIndexes []int) ([]int, error) {
	// convert outputIndexes to a set
	wantOutput := make(map[int]bool, len(outputIndexes))
	for _, outputIndex := range outputIndexes {
		wantOutput[outputIndex] = true
	}

	runningSums := make([]ColorValue, len(colorIns))
	for i, colorIn := range colorIns {
		if i == 0 {
			runningSums[i] = colorIn.ColorValue
		} else {
			runningSums[i] = runningSums[i-1] + colorIn.ColorValue
		}
	}
	inSum := runningSums[len(colorIns)-1]
	runningIndex := 0
	outSum := ColorValue(0)
	inputIndexes := make(map[int]bool, len(colorIns))
	for i, outValue := range outValues {
		cv := ColorValue(outValue - padding)
		outSum += ColorValue(cv)
		if cv <= 0 || outSum > inSum {
			continue
		}
		_, wantCurrent := wantOutput[i]
		if runningIndex < len(runningSums) && wantCurrent && colorIns[runningIndex].ColorValue != 0 {
			inputIndexes[runningIndex] = true
		}
		for runningIndex < len(runningSums) && runningSums[runningIndex] < outSum {
			runningIndex++
			if runningIndex < len(runningSums) && wantCurrent {
				inputIndexes[runningIndex] = true
			}
		}
	}
	var indexes []int
	for k := range inputIndexes {
		indexes = append(indexes, k)
	}

	sort.Ints(indexes)
	return indexes, nil
}

func (k EPOBC) FindAffectingInputs(b *BlockExplorer, genesis *btcwire.OutPoint, tx *btcwire.MsgTx, outputIndexes []int) ([]*btcwire.OutPoint, error) {

	// handle case where the tx is the issuing tx
	txShaHash, err := tx.TxSha()
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "transaction does not have a hash", err)
	}
	if genesis.Hash.IsEqual(&txShaHash) {
		return nil, nil
	}

	// calculate the input color values
	colorIns, err := k.txColorIns(b, tx)
	if err != nil {
		return nil, err
	}

	padding := k.fetchPadding(tx)
	outputValues := make([]int64, len(tx.TxOut))
	for i, out := range tx.TxOut {
		outputValues[i] = out.Value
	}
	inputIndexes, err := k.AffectingIndexes(colorIns, outputValues, padding, outputIndexes)

	var outPoints []*btcwire.OutPoint
	for _, i := range inputIndexes {
		outPoints = append(outPoints, &tx.TxIn[i].PreviousOutPoint)
	}

	return outPoints, nil
}
