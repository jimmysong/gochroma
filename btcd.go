package gochroma

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/monetas/btcnet"
	"github.com/monetas/btcrpcclient"
	"github.com/monetas/btcutil"
	"github.com/monetas/btcwire"
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
		str := fmt.Sprintf("failed to connect with params, %v", connConfig)
		return nil, MakeError(ErrConnect, str, err)
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
		str := fmt.Sprintf("failed to read at height %d", height)
		return nil, MakeError(ErrBlockRead, str, err)
	}
	return hash.Bytes(), nil
}

// RawBlock returns the raw byte-slice of the block identified by the
// byte-slice hash.
func (b *btcdBlockReaderWriter) RawBlock(hash []byte) ([]byte, error) {
	shaHash, err := btcwire.NewShaHash(hash)
	if err != nil {
		str := fmt.Sprintf("hash %x looks bad", hash)
		return nil, MakeError(ErrInvalidHash, str, err)
	}

	block, err := b.Client.GetBlock(shaHash)
	if err != nil {
		str := fmt.Sprintf("failed to get block %x", hash)
		return nil, MakeError(ErrBlockRead, str, err)
	}

	return block.Bytes()
}

// RawTx returns the raw byte-slice of the hash identified by the
// byte-slice hash.
func (b *btcdBlockReaderWriter) RawTx(hash []byte) ([]byte, error) {
	shaHash, err := btcwire.NewShaHash(hash)
	if err != nil {
		str := fmt.Sprintf("hash %x looks bad", hash)
		return nil, MakeError(ErrInvalidHash, str, err)
	}
	tx, err := b.Client.GetRawTransaction(shaHash)
	if err != nil {
		str := fmt.Sprintf("failed to get tx %x", hash)
		return nil, MakeError(ErrBlockRead, str, err)
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
		str := fmt.Sprintf("hash %x looks bad", txHash)
		return nil, MakeError(ErrInvalidHash, str, err)
	}
	txRawResult, err := b.Client.GetRawTransactionVerbose(shaHash)
	if err != nil {
		str := fmt.Sprintf("failed to get tx verbose %x", txHash)
		return nil, MakeError(ErrBlockRead, str, err)
	}
	return hex.DecodeString(txRawResult.BlockHash)
}

// MempoolTxs returns the list of transaction hashes in the mempool.
func (b *btcdBlockReaderWriter) MempoolTxs() ([][]byte, error) {
	txs, err := b.Client.GetRawMempool()
	if err != nil {
		str := fmt.Sprintf("failed to get mempool txs")
		return nil, MakeError(ErrBlockRead, str, err)
	}
	ret := make([][]byte, len(txs))
	for i, shaHash := range txs {
		ret[i] = shaHash.Bytes()
	}
	return ret, nil
}

// TxOutSpent returns a pointer to a boolean about whether an outpoint
// has been spent or not.
func (b *btcdBlockReaderWriter) TxOutSpent(hash []byte,
	index uint32, mempool bool) (*bool, error) {
	shaHash, err := btcwire.NewShaHash(hash)
	if err != nil {
		str := fmt.Sprintf("hash %x looks bad", hash)
		return nil, MakeError(ErrInvalidHash, str, err)
	}

	txOutInfo, err := b.Client.GetTxOut(shaHash, int(index), mempool)
	if err != nil {
		str := fmt.Sprintf("failed to get tx out info %x", hash)
		return nil, MakeError(ErrBlockRead, str, err)
	}

	spent := txOutInfo == nil

	return &spent, nil
}

// PublishRawTx sends the transaction to the blockchain and returns
// the byte-slice transaction id/hash.
func (b *btcdBlockReaderWriter) PublishRawTx(hash []byte) ([]byte, error) {
	tx, err := btcutil.NewTxFromBytes(hash)
	if err != nil {
		str := fmt.Sprintf("failed to convert to tx %x", hash)
		return nil, MakeError(ErrInvalidTx, str, err)
	}
	shaHash, err := b.Client.SendRawTransaction(tx.MsgTx(), true)
	if err != nil {
		str := fmt.Sprintf("failed to publish tx %x", hash)
		return nil, MakeError(ErrBlockWrite, str, err)
	}
	return shaHash.Bytes(), nil
}
