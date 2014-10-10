package gochroma

import (
	"bytes"
	"encoding/hex"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

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

// BlockCount returns the height of the newest block.
func (b *btcdBlockReaderWriter) BlockCount() (int64, error) {
	return b.Client.GetBlockCount()
}

// BlockHash returns the byte-slice hash of the block at height given.
func (b *btcdBlockReaderWriter) BlockHash(height int64) ([]byte, error) {
	hash, err := b.Client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	return hash.Bytes(), nil
}

// RawBlock returns the raw byte-slice of the block identified by the
// byte-slice hash.
func (b *btcdBlockReaderWriter) RawBlock(hash []byte) ([]byte, error) {
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

// RawTx returns the raw byte-slice of the hash identified by the
// byte-slice hash.
func (b *btcdBlockReaderWriter) RawTx(hash []byte) ([]byte, error) {
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

// TxBlockHash returns the byte-slice block hash identified by the
// byte-slice transaction hash.
func (b *btcdBlockReaderWriter) TxBlockHash(txHash []byte) ([]byte, error) {
	shaHash, err := btcwire.NewShaHash(txHash)
	if err != nil {
		return nil, err
	}
	txRawResult, err := b.Client.GetRawTransactionVerbose(shaHash)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(txRawResult.BlockHash)
}

// MempoolTxs returns the list of transaction hashes in the mempool.
func (b *btcdBlockReaderWriter) MempoolTxs() ([][]byte, error) {
	txs, err := b.Client.GetRawMempool()
	if err != nil {
		return nil, err
	}
	ret := make([][]byte, len(txs))
	for i, shaHash := range txs {
		ret[i] = shaHash.Bytes()
	}
	return ret, nil
}

// PublishRawTx sends the transaction to the blockchain and returns
// the byte-slice transaction id/hash.
func (b *btcdBlockReaderWriter) PublishRawTx(hash []byte) ([]byte, error) {
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
