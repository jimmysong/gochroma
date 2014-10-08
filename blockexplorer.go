package gochroma

import (
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

// BlockReaderWriter is any place where we can get raw blockchain data
// and publish raw blockchain data.
type BlockReaderWriter interface {
	// Get the current block height
	GetBlockCount() (int64, error)
	// Get the hash of the block with index
	GetBlockHash(index int64) ([]byte, error)
	// Get the actual block with a given hash
	GetRawBlock(hash []byte) ([]byte, error)
	// Get the raw transaction given a hash
	GetRawTx(hash []byte) ([]byte, error)
	// Get the raw transaction in the mempool
	GetMempoolTxs() ([][]byte, error)
	// Get the block hash than contains the tx identified by the tx hash
	GetTxBlockHash(txHash []byte) ([]byte, error)
	// Publish a raw transaction
	SendRawTx(rawTx []byte) ([]byte, error)
}

// BlockExplorer is a struct with methods that return btcutil-style objects
// from the BlockReaderWriter.
type BlockExplorer struct {
	BlockReaderWriter
}

// GetLatestBlock returns the *btcutil.Block struct of the latest block
// we have.
func (b *BlockExplorer) GetLatestBlock() (*btcutil.Block, error) {
	height, err := b.GetBlockCount()
	if err != nil {
		return nil, err
	}
	return b.GetBlockAtHeight(height)
}

// GetRawBlockAtHeight returns a byte slice representing the raw blocks
// that get passed around the blockchain given a height.
func (b *BlockExplorer) GetRawBlockAtHeight(height int64) ([]byte, error) {
	hash, err := b.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	return b.GetRawBlock(hash)
}

// GetBlockAtHeight returns a *btcutil.Block struct given a height.
func (b *BlockExplorer) GetBlockAtHeight(height int64) (*btcutil.Block, error) {
	raw, err := b.GetRawBlockAtHeight(height)
	if err != nil {
		return nil, err
	}
	return btcutil.NewBlockFromBytes(raw)
}

// GetBlock returns a *btcutil.Block struct given a byte-slice hash.
func (b *BlockExplorer) GetBlock(hash []byte) (*btcutil.Block, error) {
	raw, err := b.GetRawBlock(hash)
	if err != nil {
		return nil, err
	}
	return btcutil.NewBlockFromBytes(raw)
}

// GetPreviousBlock returns the *btcutil.Block struct of the block before
// a given block identified by the byte-slice hash.
func (b *BlockExplorer) GetPreviousBlock(hash []byte) (*btcutil.Block, error) {
	block, err := b.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	previousHash := block.MsgBlock().Header.PrevBlock
	return b.GetBlock(previousHash.Bytes())
}

// GetTx returns the *btcutil.Tx struct of the transaction identified by
// the byte-slice hash.
func (b *BlockExplorer) GetTx(hash []byte) (*btcutil.Tx, error) {
	raw, err := b.GetRawTx(hash)
	if err != nil {
		return nil, err
	}
	return btcutil.NewTxFromBytes(raw)
}

// GetOutPointValue returns how much many satoshis exist at this tx/index
func (b *BlockExplorer) GetOutPointValue(outpoint *btcwire.OutPoint) (int64, error) {
	tx, err := b.GetTx(outpoint.Hash.Bytes())
	if err != nil {
		return -1, err
	}
	out := tx.MsgTx().TxOut[outpoint.Index]

	return out.Value, nil
}

// GetTxBlock returns the *btcutil.Block struct of the transaction identified by
// the byte-slice hash.
func (b *BlockExplorer) GetTxBlock(txHash []byte) (*btcutil.Block, error) {
	blockHash, err := b.GetTxBlockHash(txHash)
	if err != nil {
		return nil, err
	}
	return b.GetBlock(blockHash)
}
