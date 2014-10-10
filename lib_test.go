package gochroma_test

import (
	"github.com/jimmysong/gochroma"
)

// Test block read/writer that returns whatever you initialize with.
// Note each slice is a queue that shifts one off and throws an error
// if there's none left.
type TstBlockReaderWriter struct {
	blockCount  []int64
	blockHash   [][]byte
	block       [][]byte
	rawTx       [][]byte
	txBlockHash [][]byte
	mempoolTxs  [][][]byte
	sendHash    [][]byte
}

func (b *TstBlockReaderWriter) BlockCount() (int64, error) {
	if len(b.blockCount) == 0 {
		return 0, gochroma.MakeError(gochroma.ErrBlockRead, "Block Error", nil)
	}
	ret := b.blockCount[0]
	b.blockCount = b.blockCount[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) BlockHash(_ int64) ([]byte, error) {
	if len(b.blockHash) == 0 {
		return nil, gochroma.MakeError(gochroma.ErrBlockRead, "BlockHash Error", nil)
	}
	ret := b.blockHash[0]
	b.blockHash = b.blockHash[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) RawBlock(_ []byte) ([]byte, error) {
	if len(b.block) == 0 {
		return nil, gochroma.MakeError(gochroma.ErrBlockRead, "RawBlock Error", nil)
	}
	ret := b.block[0]
	b.block = b.block[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) RawTx(_ []byte) ([]byte, error) {
	if len(b.rawTx) == 0 {
		return nil, gochroma.MakeError(gochroma.ErrBlockRead, "RawTx Error", nil)
	}
	ret := b.rawTx[0]
	b.rawTx = b.rawTx[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) TxBlockHash(_ []byte) ([]byte, error) {
	if len(b.txBlockHash) == 0 {
		return nil, gochroma.MakeError(gochroma.ErrBlockRead, "TxBlockHash Error", nil)
	}
	ret := b.txBlockHash[0]
	b.txBlockHash = b.txBlockHash[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) MempoolTxs() ([][]byte, error) {
	if len(b.mempoolTxs) == 0 {
		return nil, gochroma.MakeError(gochroma.ErrBlockRead, "MempoolTxs Error", nil)
	}
	ret := b.mempoolTxs[0]
	b.mempoolTxs = b.mempoolTxs[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) PublishRawTx(_ []byte) ([]byte, error) {
	if len(b.sendHash) == 0 {
		return nil, gochroma.MakeError(gochroma.ErrBlockWrite, "PublishRawTx Error", nil)
	}
	ret := b.sendHash[0]
	b.sendHash = b.sendHash[1:]
	return ret, nil
}
