package jsonrpc

import (
	"fmt"
)

type ErrorCode int

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

const (
	ErrCodeParseError          = -32700
	ErrCodeInvalidRequest      = -32600
	ErrCodeMethodNotFound      = -32601
	ErrCodeInvalidParams       = -32602
	ErrCodeInternalError       = -32603
	ErrCodeInvalidInput        = -32000
	ErrCodeResourceNotFound    = -32001
	ErrCodeResourceUnavailable = -32002
	ErrCodeTransactionRejected = -32003
	ErrCodeMethodNotSupported  = -32004
	ErrCodeLimitExceeded       = -32005
)

func (e *Error) Error() string {
	return e.Message
}

func ParseError(message string) *Error {
	return &Error{
		Code:    ErrCodeParseError,
		Message: message,
	}
}

func InvalidRequest(message string) *Error {
	return &Error{
		Code:    ErrCodeInvalidRequest,
		Message: message,
	}
}

func MethodNotFound(request *Request) *Error {
	return &Error{
		Code:    ErrCodeMethodNotFound,
		Message: fmt.Sprintf("The method %s does not exist/is not available", request.Method),
	}
}

func InvalidParams(message string) *Error {
	return &Error{
		Code:    ErrCodeInvalidParams,
		Message: message,
	}
}

func InternalError(message string) *Error {
	return &Error{
		Code:    ErrCodeInternalError,
		Message: message,
	}
}

func InvalidInput(message string) *Error {
	return &Error{
		Code:    ErrCodeInvalidInput,
		Message: message,
	}
}

func ResourceNotFound(message string) *Error {
	return &Error{
		Code:    ErrCodeResourceNotFound,
		Message: message,
	}
}

func ResourceUnavailable(message string) *Error {
	return &Error{
		Code:    ErrCodeResourceUnavailable,
		Message: message,
	}
}

func TransactionRejected(message string) *Error {
	return &Error{
		Code:    ErrCodeTransactionRejected,
		Message: message,
	}
}

func MethodNotSupported(request *Request) *Error {
	return &Error{
		Code:    ErrCodeMethodNotSupported,
		Message: fmt.Sprintf("method not supported %s", request.Method),
	}
}

func LimitExceeded(message string) *Error {
	return &Error{
		Code:    ErrCodeLimitExceeded,
		Message: message,
	}
}
