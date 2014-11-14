package gochroma

type BitList []bool

func NewBitList(number uint32, bits int) BitList {
	b := make([]bool, bits)
	for i := range b {
		b[i] = (number>>uint(i))&1 == 1
	}
	return BitList(b)
}

func (b BitList) Uint32() uint32 {
	number := uint32(0)
	current := uint32(1)
	for _, bit := range b {
		if bit {
			number += current
		}
		current *= 2
	}
	return number
}

func (b BitList) Combine(b2 BitList) BitList {
	return append(b, b2...)
}

func (b BitList) Equal(b2 BitList) bool {
	if len(b) != len(b2) {
		return false
	}
	for i := range b {
		if b[i] != b2[i] {
			return false
		}
	}
	return true
}
