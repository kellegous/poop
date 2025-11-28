package poop

import (
	"bytes"
	"errors"
	"fmt"
	"iter"
)

type chainedError struct {
	caller
	message string
	next    error
}

func (e *chainedError) Error() string {
	for err := range IterChain(e) {
		if cerr, ok := err.(*chainedError); ok {
			if m := cerr.message; m != "" {
				return m
			}
		} else {
			return err.Error()
		}
	}
	panic("nil error chain")
}

func (e *chainedError) Unwrap() error {
	return e.next
}

func newChainedError(
	next error,
	message string,
	caller caller,
) error {
	return &chainedError{
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
		if cerr, ok := e.(*chainedError); ok {
			f := cerr.frame()
			buf.WriteString(fmt.Sprintf("%s(%s:%d)", f.function, pf(f.file), f.line))
			if m := cerr.message; m != "" {
				buf.WriteString(fmt.Sprintf(" %s", m))
			}
		} else {
			buf.WriteString(e.Error())
		}
	}

	return errors.New(buf.String())
}
