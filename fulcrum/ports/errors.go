package ports

import "errors"

var (
	ErrorSystemFailure  = errors.New("system failure") // Base error
	ErrorIncorrectInput = errors.New("incorrect input")
)
