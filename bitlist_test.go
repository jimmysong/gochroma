package gochroma_test

import (
	"testing"

	"github.com/monetas/gochroma"
)

func TestNewBitList(t *testing.T) {
	// Execute
	bits := gochroma.NewBitList(13, 4)

	// Validate
	want := gochroma.BitList{true, false, true, true}
	if !bits.Equal(want) {
		t.Fatalf("unexpected bits: got %v, want %v", bits, want)
	}
}

func TestBitListEqual(t *testing.T) {
	tests := []struct {
		a     gochroma.BitList
		b     gochroma.BitList
		equal bool
	}{
		{
			a:     gochroma.BitList{true},
			b:     gochroma.BitList{true},
			equal: true,
		},
		{
			a:     gochroma.BitList{false},
			b:     gochroma.BitList{true},
			equal: false,
		},
		{
			a:     gochroma.BitList{false, true},
			b:     gochroma.BitList{false},
			equal: false,
		},
	}
	for _, test := range tests {
		if test.a.Equal(test.b) != test.equal {
			t.Fatalf("should be %v equality: a=%v, b=%v", test.equal, test.a, test.b)
		}
	}
}

func TestUint32(t *testing.T) {
	// Setup
	want := uint32(31)
	bits := gochroma.NewBitList(want, 32)

	// Execute
	got := bits.Uint32()

	// Validate
	if got != want {
		t.Fatalf("unexpected uint32: got %v, want %v", got, want)
	}
}

func TestCombine(t *testing.T) {
	// Setup
	b1 := gochroma.NewBitList(10, 4)
	b2 := gochroma.NewBitList(5, 4)
	want := gochroma.NewBitList(90, 8)

	// Execute
	got := b1.Combine(b2)

	if !got.Equal(want) {
		t.Fatalf("unexpected combine: got %v, want %v", got, want)
	}
}
