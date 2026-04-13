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

func (h *Handler) logError(err error, context string) {
	if h.logger != nil {
		h.logger.Error(context, "error", err)
	}
}

func (h *Handler) logWarning(msg string, args ...any) {
	if h.logger != nil {
		h.logger.Warn(msg, args...)
	}
}

func (h *Handler) logInfo(msg string, args ...any) {
	if h.logger != nil {
		h.logger.Info(msg, args...)
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
