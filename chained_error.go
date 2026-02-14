package poop

import (
	"bytes"
	"errors"
	"fmt"
	"iter"
)

// ChainedError represents a link in the causal chain of errors.
type ChainedError struct {
	caller
	message string
	next    error
}

func (e *ChainedError) Error() string {
	for err := range IterChain(e) {
		if cerr, ok := err.(*ChainedError); ok {
			if m := cerr.message; m != "" {
				return m
			}
		} else {
			return err.Error()
		}
	}
	panic("nil error chain")
}

func (e *ChainedError) Unwrap() error {
	return e.next
}

// Frame returns the frame information for the caller of the error.
// This will be the call frame for the chain link.
func (e *ChainedError) Frame() Frame {
	return e.caller.frame()
}

// Message returns the message that was provided when the error
// was chained. The message is optional and may be empty.
func (e *ChainedError) Message() string {
	return e.message
}

// RootCause returns the first error in the chain.
func (e *ChainedError) RootCause() error {
	var err error
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
}

func newChainedError(
	next error,
	message string,
	caller caller,
) error {
	return &ChainedError{
		caller:  caller,
		message: message,
		next:    next,
	}
}

// New creates a leaf error with caller information. This is the poop equivalent to `errors.New`.
func New(message string) error {
	return newChainedError(nil, message, callerFunc())
}

// Newf is identical to New, but allows formatted messages.
func Newf(format string, args ...interface{}) error {
	return newChainedError(nil, fmt.Sprintf(format, args...), callerFunc())
}

// Chain chains the given error. This is the most common way to chain. It captures the
// caller information, the location where this error is being returned, but doesn't require
// any other information.
func Chain(err error) error {
	if err == nil {
		return nil
	}
	return newChainedError(err, "", callerFunc())
}

// ChainWith chains the given error with an additional message.
func ChainWith(err error, message string) error {
	if err == nil {
		return nil
	}
	return newChainedError(err, message, callerFunc())
}

// ChainWithf is identical to ChainWith, but allows formatted messages.
func ChainWithf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return newChainedError(err, fmt.Sprintf(format, args...), callerFunc())
}

// IterChain is an iterator of all the errors in the chain.
//
// TODO(kellegous): Since ChainError is not exported, this should
// be changed to a method on ChainedError.
func IterChain(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		for {
			if !yield(err) {
				return
			}

			err = errors.Unwrap(err)
			if err == nil {
				return
			}
		}
	}
}

// Flatten flattens the chain into a single error with the information
// about the chain in the error message. This is useful in contexts where
// you need to capture everyting in the string returned via Error, like zap
// logging, zap.Error(poop.Flatten(err)).
func Flatten(err error) error {
	pf := PathLastNSegments(2)

	var buf bytes.Buffer
	for e := range IterChain(err) {
		if buf.Len() > 0 {
			buf.WriteString(" → ")
		}
		if cerr, ok := e.(*ChainedError); ok {
			f := cerr.frame()
			buf.WriteString(fmt.Sprintf("%s(%s:%d)", f.Function, pf(f.File), f.Line))
			if m := cerr.message; m != "" {
				buf.WriteString(fmt.Sprintf(" %s", m))
			}
		} else {
			buf.WriteString(e.Error())
		}
	}

	return errors.New(buf.String())
}
