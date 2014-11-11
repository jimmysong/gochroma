package gochroma_test

import (
	"testing"

	"github.com/jimmysong/gochroma"
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
	// Setup
	got := gochroma.BitList{true}
	want := gochroma.BitList{true}

	// Execute
	// Validate
	if !got.Equal(want) {
		t.Fatalf("unexpected bits: got %v, want %v", got, want)
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
