package gochroma

func MakeError(c ErrorCode, d string, e error) ChromaError {
	return makeError(c, d, e)
}
