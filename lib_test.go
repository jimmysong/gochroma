package gochroma_test

import (
	"errors"
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
		return 0, errors.New("Block Error")
	}
	ret := b.blockCount[0]
	b.blockCount = b.blockCount[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) BlockHash(_ int64) ([]byte, error) {
	if len(b.blockHash) == 0 {
		return nil, errors.New("BlockHash Error")
	}
	ret := b.blockHash[0]
	b.blockHash = b.blockHash[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) RawBlock(_ []byte) ([]byte, error) {
	if len(b.block) == 0 {
		return nil, errors.New("RawBlock Error")
	}
	ret := b.block[0]
	b.block = b.block[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) RawTx(_ []byte) ([]byte, error) {
	if len(b.rawTx) == 0 {
		return nil, errors.New("RawTx Error")
	}
	ret := b.rawTx[0]
	b.rawTx = b.rawTx[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) TxBlockHash(_ []byte) ([]byte, error) {
	if len(b.txBlockHash) == 0 {
		return nil, errors.New("TxBlockHash Error")
	}
	ret := b.txBlockHash[0]
	b.txBlockHash = b.txBlockHash[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) MempoolTxs() ([][]byte, error) {
	if len(b.mempoolTxs) == 0 {
		return nil, errors.New("MempoolTxs Error")
	}
	ret := b.mempoolTxs[0]
	b.mempoolTxs = b.mempoolTxs[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) PublishRawTx(_ []byte) ([]byte, error) {
	if len(b.sendHash) == 0 {
		return nil, errors.New("PublishRawTx Error")
	}
	ret := b.sendHash[0]
	b.sendHash = b.sendHash[1:]
	return ret, nil
}
