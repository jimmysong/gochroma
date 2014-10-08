package gochroma_test

import (
	"errors"
	"testing"

	"github.com/jimmysong/gochroma"
)

func TestString(t *testing.T) {
	// Setup
	e := gochroma.ErrorCode(gochroma.ErrUnimplemented)

	// Execute
	s := e.String()

	// Verify
	wantStr := "unimplemented"
	if s != wantStr {
		t.Fatalf("wrong error string: got %v want %v", s, wantStr)
	}
}

func TestStringUnknown(t *testing.T) {
	// Setup
	e := gochroma.ErrorCode(-1)

	// Execute
	s := e.String()

	// Verify
	wantStr := "Unknown ErrorCode: -1"
	if s != wantStr {
		t.Fatalf("wrong error string: got %v want %v", s, wantStr)
	}
}

func TestError1(t *testing.T) {
	// Setup
	wantStr := "test"
	e := gochroma.MakeError(gochroma.ErrUnimplemented, wantStr, nil)

	// Execute
	s := e.Error()

	// Verify
	if s != wantStr {
		t.Fatalf("wrong error string: got %v want %v", s, wantStr)
	}
}

func TestError2(t *testing.T) {
	// Setup
	wantStr := "test: test2"
	e := gochroma.MakeError(gochroma.ErrUnimplemented, "test", errors.New("test2"))

	// Execute
	s := e.Error()

	// Verify
	if s != wantStr {
		t.Fatalf("wrong error string: got %v want %v", s, wantStr)
	}
}
