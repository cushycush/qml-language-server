package handler

import (
	"errors"
	"fmt"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrParserNotReady   = errors.New("parser not initialized")
	ErrInvalidPosition  = errors.New("invalid document position")
	ErrTreeNotAvailable = errors.New("parse tree not available")
)

type HandlerError struct {
	Code    string
	Message string
	Err     error
}

func (e *HandlerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *HandlerError) Unwrap() error {
	return e.Err
}

func newHandlerError(code, message string, err error) *HandlerError {
	return &HandlerError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func safeString(s string) string {
	if s == "" {
		return "<empty>"
	}
	return s
}

func safeSliceLen[T any](s []T) int {
	if s == nil {
		return 0
	}
	return len(s)
}
