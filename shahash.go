package gochroma

import (
	"github.com/conformal/btcwire"
)

// reverse simply reverses the bytes
func reverse(input []byte) []byte {
	ret := make([]byte, len(input))
	copy(ret, input)
	for i, j := 0, len(ret)-1; i < j; i, j = i+1, j-1 {
		ret[i], ret[j] = ret[j], ret[i]
	}
	return ret
}

// NewShaHash takes a big-endian bytes and returns the ShaHash that
// corresponds. We use this because the btcwire NewShaHash assumes
// the bytes are little-endian
func NewShaHash(bigEndianBytes []byte) (*btcwire.ShaHash, error) {
	littleEndianBytes := reverse(bigEndianBytes)
	return btcwire.NewShaHash(littleEndianBytes)
}

// BigEndianBytes returns the bytes in big-endian order
func BigEndianBytes(hash *btcwire.ShaHash) []byte {
	return reverse(hash.Bytes())
}
