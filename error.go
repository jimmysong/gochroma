package gochroma

import (
	"fmt"
)

const (
	ErrUnimplemented = iota
	ErrInvalidTx
	ErrTooMuchColorValue
	ErrTooManyOutputs
	ErrBadOutputIndex
	ErrInvalidColorValue
	ErrUnknownKernel
)

type ErrorCode int

var errCodeStrings = map[ErrorCode]string{
	ErrUnimplemented:     "unimplemented",
	ErrInvalidTx:         "transaction is invalid",
	ErrTooMuchColorValue: "too much color value",
	ErrTooManyOutputs:    "too many outputs",
	ErrBadOutputIndex:    "output index is bad",
	ErrInvalidColorValue: "ColorValue is invalid",
	ErrUnknownKernel:     "unknown kernel",
}

func (e ErrorCode) String() string {
	s, ok := errCodeStrings[e]
	if ok {
		return s
	} else {
		return fmt.Sprintf("Unknown ErrorCode: %d", int(e))
	}
}

type ChromaError struct {
	ErrorCode   ErrorCode
	Description string
	Err         error
}

func (e ChromaError) Error() string {
	if e.Err != nil {
		return e.Description + ": " + e.Err.Error()
	}
	return e.Description
}

func makeError(c ErrorCode, d string, e error) ChromaError {
	return ChromaError{
		ErrorCode:   c,
		Description: d,
		Err:         e,
	}
}
