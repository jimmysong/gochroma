package gochroma

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/monetas/btcwire"
	"github.com/monetas/fastsha256"
)

type ColorValue uint64

type ColorIn struct {
	OutPoint   *btcwire.OutPoint
	ColorValue ColorValue
}

func OutPoints(cis []*ColorIn) []*btcwire.OutPoint {
	ret := make([]*btcwire.OutPoint, len(cis))
	for i, ci := range cis {
		ret[i] = ci.OutPoint
	}
	return ret
}

type ColorOut struct {
	Script     []byte
	ColorValue ColorValue
}

type ColorKernel interface {
	// 4-6 letter code for the kernel
	Code() string
	// Takes any outpoint and determines the color value given the genesis
	OutPointToColorIn(b *BlockExplorer, genesis, outPoint *btcwire.OutPoint) (*ColorIn, error)
	// Returns the number of satoshi needed for a particular color value
	IssuingSatoshiNeeded(cv ColorValue) int64
	// Validates the color inputs and checks if the color values
	// correspond to this kernel and genesis
	ColorInsValid(b *BlockExplorer, genesis *btcwire.OutPoint, colorIns []*ColorIn) (bool, error)
	// Returns the unsigned transaction that will issue the color
	// of this kernel with a certain color value
	IssuingTx(b *BlockExplorer, inputs []*btcwire.OutPoint, outputs []*ColorOut, changeScript []byte, fee int64) (*btcwire.MsgTx, error)
	// Returns the unsigned transaction that will transfer the color values
	// to the desired places
	TransferringTx(b *BlockExplorer, inputs []*ColorIn, outputs []*ColorOut, changeScript []byte, fee int64, destroy bool) (*btcwire.MsgTx, error)
	// Calculates the output color values given the input color values
	// based on the kernel rules.
	CalculateOutColorValues(genesis *btcwire.OutPoint, tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error)
	// Figures out which inputs the outputs were affected by.
	// Note the outputs array is the collection of indices for tx.TxOuts
	FindAffectingInputs(b *BlockExplorer, genesis *btcwire.OutPoint, tx *btcwire.MsgTx, outputIndexes []int) ([]*btcwire.OutPoint, error)
}

var kernelMap = make(map[string]ColorKernel, 10)

func RegisterColorKernel(kernel ColorKernel) error {
	key := kernel.Code()
	_, ok := kernelMap[key]
	if ok {
		// this is a duplicate
		str := fmt.Sprintf("%v is already a registered kernel", key)
		return MakeError(ErrDuplicateKernel, str, nil)
	}
	kernelMap[key] = kernel
	return nil
}

func GetColorKernel(key string) (ColorKernel, error) {
	kernel, ok := kernelMap[key]
	if !ok {
		str := fmt.Sprintf("%v is not a registered kernel", key)
		return nil, MakeError(ErrNonExistentKernel, str, nil)
	}
	return kernel, nil
}

type ColorDefinition struct {
	ColorKernel
	Genesis *btcwire.OutPoint
	Height  int64
}

func (c *ColorDefinition) String() string {
	return fmt.Sprintf("%v:%v:%d:%d", c.Code(), c.Genesis.Hash, c.Genesis.Index, c.Height)
}

func (c *ColorDefinition) HashString() string {
	return fmt.Sprintf("%v:%v:%d", c.Code(), c.Genesis.Hash, c.Genesis.Index)
}

func (c *ColorDefinition) Hash() []byte {
	hash := fastsha256.Sum256([]byte(c.HashString()))
	return hash[:]
}

func (c *ColorDefinition) AccountNumber() uint32 {
	cdHash := c.Hash()
	return binary.LittleEndian.Uint32(cdHash[:4]) % (1 << 31)
}

func (c *ColorDefinition) RunKernel(tx *btcwire.MsgTx, inputs []ColorValue) ([]ColorValue, error) {
	return c.CalculateOutColorValues(c.Genesis, tx, inputs)
}

func (c *ColorDefinition) AffectingInputs(b *BlockExplorer, tx *btcwire.MsgTx, outputIndexes []int) ([]*btcwire.OutPoint, error) {
	return c.FindAffectingInputs(b, c.Genesis, tx, outputIndexes)
}

func (c *ColorDefinition) ColorValue(b *BlockExplorer, outPoint *btcwire.OutPoint) (*ColorValue, error) {
	colorIn, err := c.OutPointToColorIn(b, c.Genesis, outPoint)
	if err != nil {
		return nil, err
	}
	return &colorIn.ColorValue, nil
}

func NewColorDefinition(kernel ColorKernel, genesis *btcwire.OutPoint, height int64) (*ColorDefinition, error) {
	return &ColorDefinition{
		kernel, genesis, height,
	}, nil
}

func NewColorDefinitionFromStr(cdString string) (*ColorDefinition, error) {
	components := strings.Split(cdString, ":")
	if len(components) != 4 {
		return nil, MakeError(ErrBadColorDefinition, "color definition should have 4 components", nil)
	}
	kernel, err := GetColorKernel(components[0])
	if err != nil {
		return nil, err
	}
	shaHash, err := btcwire.NewShaHashFromStr(components[1])
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "hash is invalid", err)
	}
	index, err := strconv.Atoi(components[2])
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "index is invalid", err)
	}
	genesis := btcwire.NewOutPoint(shaHash, uint32(index))

	height, err := strconv.Atoi(components[3])
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "height is invalid", err)
	}
	if height < 0 {
		return nil, MakeError(ErrInvalidTx, "height is negative", nil)
	}

	return &ColorDefinition{
		kernel, genesis, int64(height),
	}, nil
}
