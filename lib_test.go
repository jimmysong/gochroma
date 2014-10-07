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

func (b *TstBlockReaderWriter) GetBlockCount() (int64, error) {
	if len(b.blockCount) == 0 {
		return 0, errors.New("GetBlock Error")
	}
	ret := b.blockCount[0]
	b.blockCount = b.blockCount[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) GetBlockHash(_ int64) ([]byte, error) {
	if len(b.blockHash) == 0 {
		return nil, errors.New("GetBlockHash Error")
	}
	ret := b.blockHash[0]
	b.blockHash = b.blockHash[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) GetRawBlock(_ []byte) ([]byte, error) {
	if len(b.block) == 0 {
		return nil, errors.New("GetRawBlock Error")
	}
	ret := b.block[0]
	b.block = b.block[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) GetRawTx(_ []byte) ([]byte, error) {
	if len(b.rawTx) == 0 {
		return nil, errors.New("GetRawTx Error")
	}
	ret := b.rawTx[0]
	b.rawTx = b.rawTx[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) GetTxBlockHash(_ []byte) ([]byte, error) {
	if len(b.txBlockHash) == 0 {
		return nil, errors.New("GetTxBlockHash Error")
	}
	ret := b.txBlockHash[0]
	b.txBlockHash = b.txBlockHash[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) GetMempoolTxs() ([][]byte, error) {
	if len(b.mempoolTxs) == 0 {
		return nil, errors.New("GetMempoolTxs Error")
	}
	ret := b.mempoolTxs[0]
	b.mempoolTxs = b.mempoolTxs[1:]
	return ret, nil
}

func (b *TstBlockReaderWriter) SendRawTx(_ []byte) ([]byte, error) {
	if len(b.sendHash) == 0 {
		return nil, errors.New("SendRawTx Error")
	}
	ret := b.sendHash[0]
	b.sendHash = b.sendHash[1:]
	return ret, nil
}
