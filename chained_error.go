package poop

import (
	"bytes"
	"errors"
	"fmt"
	"iter"
)

type chainedError struct {
	Caller
	Current error
	next    error
}

func (e *chainedError) Error() string {
	for err := range IterChain(e) {
		if cerr, ok := err.(*chainedError); ok {
			if c := cerr.Current; c != nil {
				return c.Error()
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
	current error,
	caller Caller,
) error {
	return &chainedError{
		Caller:  caller,
		Current: current,
		next:    next,
	}
}

func isChainedError(err error) bool {
	_, ok := err.(*chainedError)
	return ok
}

func New(message string) error {
	return newChainedError(nil, errors.New(message), callerFunc())
}

func Newf(format string, args ...interface{}) error {
	return newChainedError(nil, fmt.Errorf(format, args...), callerFunc())
}

func Chain(err error) error {
	if err == nil {
		return nil
	} else if !isChainedError(err) {
		// we don't have caller information on the original error,
		// so we're just going to give the current caller information
		// to the top of the chain.
		return newChainedError(errors.Unwrap(err), err, callerFunc())
	}
	return newChainedError(err, nil, callerFunc())
}

func ChainWith(err error, message string) error {
	if err == nil {
		return nil
	}
	return newChainedError(err, errors.New(message), callerFunc())
}

func ChainWithf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return newChainedError(err, fmt.Errorf(format, args...), callerFunc())
}

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
			buf.WriteString(fmt.Sprintf("(%s:%d)", pf(cerr.File), cerr.Line))
			if cerr.Current != nil {
				buf.WriteString(fmt.Sprintf(" %s", cerr.Current.Error()))
			}
		} else {
			buf.WriteString(e.Error())
		}
	}

	return errors.New(buf.String())
}
