package gochroma

import (
	"bytes"

	"github.com/monetas/btcutil"
	"github.com/monetas/btcwire"
)

// BlockReaderWriter is any place where we can get raw blockchain data
// and publish raw blockchain data.
type BlockReaderWriter interface {
	// Get the current block height.
	BlockCount() (int64, error)
	// Get the hash of the block with index.
	BlockHash(index int64) ([]byte, error)
	// Get the actual block with a given hash.
	RawBlock(hash []byte) ([]byte, error)
	// Get the raw transaction given a hash.
	RawTx(hash []byte) ([]byte, error)
	// Get all the raw transactions in the mempool.
	MempoolTxs() ([][]byte, error)
	// Get the block hash that contains the tx identified by the tx hash.
	TxBlockHash(txHash []byte) ([]byte, error)
	// Get whether a transaction output is spent or not.
	TxOutSpent(txHash []byte, index uint32, mempool bool) (*bool, error)
	// Publish a raw transaction.
	PublishRawTx(rawTx []byte) ([]byte, error)
}

// BlockExplorer is a struct with methods that return btcutil-style objects
// from the BlockReaderWriter.
type BlockExplorer struct {
	BlockReaderWriter
}

// LatestBlock returns the *btcutil.Block struct of the latest block
// we have.
func (b *BlockExplorer) LatestBlock() (*btcutil.Block, error) {
	height, err := b.BlockCount()
	if err != nil {
		return nil, err
	}
	return b.BlockAtHeight(height)
}

// RawBlockAtHeight returns a byte slice representing the raw blocks
// that get passed around the blockchain given a height.
func (b *BlockExplorer) RawBlockAtHeight(height int64) ([]byte, error) {
	hash, err := b.BlockHash(height)
	if err != nil {
		return nil, err
	}
	return b.RawBlock(hash)
}

// BlockAtHeight returns a *btcutil.Block struct given a height.
func (b *BlockExplorer) BlockAtHeight(height int64) (*btcutil.Block, error) {
	raw, err := b.RawBlockAtHeight(height)
	if err != nil {
		return nil, err
	}
	return btcutil.NewBlockFromBytes(raw)
}

// Block returns a *btcutil.Block struct given a byte-slice hash.
func (b *BlockExplorer) Block(hash []byte) (*btcutil.Block, error) {
	raw, err := b.RawBlock(hash)
	if err != nil {
		return nil, err
	}
	return btcutil.NewBlockFromBytes(raw)
}

// PreviousBlock returns the *btcutil.Block struct of the block before
// a given block identified by the byte-slice hash.
func (b *BlockExplorer) PreviousBlock(hash []byte) (*btcutil.Block, error) {
	block, err := b.Block(hash)
	if err != nil {
		return nil, err
	}
	previousHash := block.MsgBlock().Header.PrevBlock
	return b.Block(previousHash.Bytes())
}

// Tx returns the *btcutil.Tx struct of the transaction identified by
// the byte-slice hash.
func (b *BlockExplorer) Tx(hash []byte) (*btcutil.Tx, error) {
	raw, err := b.RawTx(hash)
	if err != nil {
		return nil, err
	}
	return btcutil.NewTxFromBytes(raw)
}

// TxBlock returns the *btcutil.Block struct of the transaction identified by
// the byte-slice hash.
func (b *BlockExplorer) TxBlock(txHash []byte) (*btcutil.Block, error) {
	blockHash, err := b.TxBlockHash(txHash)
	if err != nil {
		return nil, err
	}
	return b.Block(blockHash)
}

// TxHeight returns the height of the block that contains the tx
func (b *BlockExplorer) TxHeight(txHash []byte) (int64, error) {
	block, err := b.TxBlock(txHash)
	if err != nil {
		return -1, err
	}
	return block.Height(), nil
}

// OutPointValue returns how much many satoshis exist at this tx/index
func (b *BlockExplorer) OutPointValue(outpoint *btcwire.OutPoint) (int64, error) {
	tx, err := b.OutPointTx(outpoint)
	if err != nil {
		return -1, err
	}
	return tx.MsgTx().TxOut[outpoint.Index].Value, nil
}

// OutPointTx returns the transaction the outpoint points to
func (b *BlockExplorer) OutPointTx(outpoint *btcwire.OutPoint) (*btcutil.Tx, error) {
	// outpoint is a shaHash, which means we have to change it to convert
	// to big-endian
	return b.Tx(BigEndianBytes(&outpoint.Hash))
}

// OutPointTx returns the transaction the outpoint points to
func (b *BlockExplorer) OutPointHeight(outpoint *btcwire.OutPoint) (int64, error) {
	// outpoint is a shaHash, which means we have to change it to convert
	// to big-endian
	return b.TxHeight(BigEndianBytes(&outpoint.Hash))
}

// OutPointSpent returns a pointer to a boolean expressing whether the outpoint
// has been spent or not.
func (b *BlockExplorer) OutPointSpent(outpoint *btcwire.OutPoint) (*bool, error) {
	return b.TxOutSpent(BigEndianBytes(&outpoint.Hash), outpoint.Index, true)
}

// PublishTx publishes the tx and then returns the shaHash of the tx.
func (b *BlockExplorer) PublishTx(tx *btcwire.MsgTx) (*btcwire.ShaHash, error) {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return nil, MakeError(ErrInvalidTx, "unable to serialize: ", err)
	}
	serialized := buffer.Bytes()
	txHash, err := b.PublishRawTx(serialized)
	if err != nil {
		return nil, err
	}
	return NewShaHash(txHash)
}
