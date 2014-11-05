package gochroma

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"

	"github.com/conformal/btcwire"
)

func SerializeUint32(i uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, i)
	return buf
}

func DeserializeUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func DeserializeOutPoint(b []byte) (*btcwire.OutPoint, error) {
	var buf bytes.Buffer
	buf.Write(b)
	dec := gob.NewDecoder(&buf)
	var op btcwire.OutPoint
	err := dec.Decode(&op)
	if err != nil {
		return nil, err
	}
	return &op, nil
}

func SerializeOutPoint(op *btcwire.OutPoint) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(op)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DeserializeColorOutPoint(b []byte) (*ColorOutPoint, error) {
	var buf bytes.Buffer
	buf.Write(b)
	dec := gob.NewDecoder(&buf)
	var cop ColorOutPoint
	err := dec.Decode(&cop)
	if err != nil {
		return nil, err
	}
	return &cop, nil
}

func SerializeColorOutPoint(cop *ColorOutPoint) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(cop)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
