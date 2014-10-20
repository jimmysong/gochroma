package gochroma

import (
	"fmt"
)

const (
	ErrUnimplemented = iota
	ErrConnect
	ErrBlockRead
	ErrBlockWrite
	ErrInvalidHash
	ErrInvalidTx
	ErrTooMuchColorValue
	ErrTooManyOutputs
	ErrBadOutputIndex
	ErrInvalidColorValue
	ErrBadColorDefinition
	ErrDuplicateKernel
	ErrNonExistentKernel
	ErrInsufficientFunds
	ErrNegativeValue
	ErrInsufficientColorValue
	ErrDestroyColorValue
	ErrUnknownKernel
)

type ErrorCode int

var errCodeStrings = map[ErrorCode]string{
	ErrUnimplemented:          "unimplemented",
	ErrConnect:                "unable to connect to blockchain source",
	ErrBlockRead:              "unable to read from blockchain source",
	ErrBlockWrite:             "unable to write to blockchain source",
	ErrInvalidHash:            "hash looks wrong",
	ErrInvalidTx:              "transaction is invalid",
	ErrTooMuchColorValue:      "too much color value",
	ErrTooManyOutputs:         "too many outputs",
	ErrBadOutputIndex:         "output index is bad",
	ErrInvalidColorValue:      "ColorValue is invalid",
	ErrBadColorDefinition:     "Color Definition is unparseable",
	ErrDuplicateKernel:        "Kernel already registered",
	ErrNonExistentKernel:      "Kernel does not exist",
	ErrInsufficientFunds:      "funds are insufficient to complete this tx",
	ErrNegativeValue:          "bitcoin amounts cannot be negative",
	ErrInsufficientColorValue: "color funds are insufficient to complete this tx",
	ErrDestroyColorValue:      "color funds are being destroyed",
	ErrUnknownKernel:          "unknown kernel",
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

func MakeError(c ErrorCode, d string, e error) ChromaError {
	return ChromaError{
		ErrorCode:   c,
		Description: d,
		Err:         e,
	}
}
