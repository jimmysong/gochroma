package gochroma

import (
	"bytes"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

// BlockExplorer represents a place where we can get blockchain data
// and publish blockchain data
type BlockExplorer interface {
	// Get the current block height
	GetBlockCount() (int64, error)
	// Get the hash of the block with index
	GetBlockHash(index int64) ([]byte, error)
	// Get the actual block with a given hash
	GetBlock(hash []byte) (*btcutil.Block, error)
	// Get the raw transaction given a hash
	GetRawTransaction(hash []byte) ([]byte, error)
	// Publish a raw transaction
	SendRawTransaction(rawTx []byte) ([]byte, error)
}

type BtcdBlockExplorer struct {
	Net    *btcnet.Params
	Client *btcrpcclient.Client
}

func NewBtcdBlockExplorer(net *btcnet.Params, connConfig *btcrpcclient.ConnConfig) (*BtcdBlockExplorer, error) {
	client, err := btcrpcclient.New(connConfig, nil)
	if err != nil {
		return nil, err
	}
	return &BtcdBlockExplorer{
		Net:    net,
		Client: client,
	}, nil
}

func (b *BtcdBlockExplorer) GetBlockCount() (int64, error) {
	return b.Client.GetBlockCount()
}

func (b *BtcdBlockExplorer) GetBlockHash(height int64) ([]byte, error) {
	hash, err := b.Client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	return hash.Bytes(), nil
}

func (b *BtcdBlockExplorer) GetBlock(hash []byte) (*btcutil.Block, error) {
	shaHash, err := btcwire.NewShaHash(hash)
	if err != nil {
		return nil, err
	}
	return b.Client.GetBlock(shaHash)
}

func (b *BtcdBlockExplorer) GetRawTransaction(hash []byte) ([]byte, error) {
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

func (b *BtcdBlockExplorer) SendRawTransaction(hash []byte) ([]byte, error) {
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
