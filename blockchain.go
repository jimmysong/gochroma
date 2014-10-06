package gochroma

import (
	"bytes"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
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
	// Publish a raw transaction
	SendRawTx(rawTx []byte) ([]byte, error)
}

// BlockExplorer is a struct with methods that return btcutil-style objects
// from the BlockReaderWriter.
type BlockExplorer struct {
	BlockReaderWriter
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

// btcdBlockReaderWriter is a specific BlockReaderWriter that uses btcd in order
// to get all the blockchain data.
type btcdBlockReaderWriter struct {
	Net    *btcnet.Params
	Client *btcrpcclient.Client
}

// NewBtcdBlockExplorer returns a BlockExplorer given a network (mainnet/
// testnet/simnet) and a connection configuration to the btcd instance.
func NewBtcdBlockExplorer(net *btcnet.Params, connConfig *btcrpcclient.ConnConfig) (*BlockExplorer, error) {
	client, err := btcrpcclient.New(connConfig, nil)
	if err != nil {
		return nil, err
	}
	return &BlockExplorer{&btcdBlockReaderWriter{
		Net:    net,
		Client: client,
	}}, nil
}

// GetBlockCount returns the height of the newest block.
func (b *btcdBlockReaderWriter) GetBlockCount() (int64, error) {
	return b.Client.GetBlockCount()
}

// GetBlockHash returns the byte-slice hash of the block at height given.
func (b *btcdBlockReaderWriter) GetBlockHash(height int64) ([]byte, error) {
	hash, err := b.Client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	return hash.Bytes(), nil
}

// GetRawBlock returns the raw byte-slice of the block identified by the
// byte-slice hash.
func (b *btcdBlockReaderWriter) GetRawBlock(hash []byte) ([]byte, error) {
	shaHash, err := btcwire.NewShaHash(hash)
	if err != nil {
		return nil, err
	}

	block, err := b.Client.GetBlock(shaHash)
	if err != nil {
		return nil, err
	}

	return block.Bytes()
}

// GetRawTx returns the raw byte-slice of the hash identified by the
// byte-slice hash.
func (b *btcdBlockReaderWriter) GetRawTx(hash []byte) ([]byte, error) {
	shaHash, err := btcwire.NewShaHash(hash)
	if err != nil {
		return nil, err
	}
	tx, err := b.Client.GetRawTransaction(shaHash)
	if err != nil {
		return nil, err
	}
	msgTx := tx.MsgTx()
	var ret bytes.Buffer
	err = msgTx.Serialize(&ret)
	if err != nil {
		return nil, err
	}
	return ret.Bytes(), nil
}

// SendRawTx sends the transaction to the blockchain and returns the byte-slice
// transaction id/hash.
func (b *btcdBlockReaderWriter) SendRawTx(hash []byte) ([]byte, error) {
	tx, err := btcutil.NewTxFromBytes(hash)
	if err != nil {
		return nil, err
	}
	shaHash, err := b.Client.SendRawTransaction(tx.MsgTx(), true)
	if err != nil {
		return nil, err
	}
	return shaHash.Bytes(), nil
}
