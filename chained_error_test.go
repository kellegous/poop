package poop

import (
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	tests := []struct {
		Name     string
		Error    error
		Expected string
	}{
		{
			"ChainedError at leaf",
			New("egad"),
			"egad",
		},
		{
			"ChainedError at leaf w/ format",
			Newf("egad %d", 1),
			"egad 1",
		},
		{
			"Unchained error at leaf",
			errors.New("egad"),
			"egad",
		},
		{
			"Single chain w/ unchained at leaf",
			Chain(errors.New("egad")),
			"egad",
		},
		{
			"Long chain w/ unchained at leaf",
			Chain(Chain(Chain(Chain(errors.New("egad"))))),
			"egad",
		},
		{
			"Long chain w/ interstitial message",
			Chain(ChainWith(Chain(errors.New("egad")), "oh noes")),
			"oh noes",
		},
		{
			"Long chain w/ interstitial message",
			Chain(ChainWithf(Chain(errors.New("egad")), "oh noes %d", 1)),
			"oh noes 1",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if test.Error.Error() != test.Expected {
				t.Fatalf("expected %q but got %q", test.Expected, test.Error.Error())
			}
		})
	}
}

func newSerialCaller() func() caller {
	i := 0
	return func() caller {
		i++
		return caller{
			File: fmt.Sprintf("file-%d", i),
			Line: i,
		}
	}
}

func isSameError(a, b error) bool {
	if a == b {
		return true
	} else if a == nil || b == nil {
		return false
	}

	ca, aok := a.(*chainedError)
	cb, bok := b.(*chainedError)
	if aok != bok {
		return false
	} else if aok {
		return isSameError(ca.current, cb.current) &&
			isSameError(ca.next, cb.next) &&
			ca.caller == cb.caller
	}

	return a.Error() == b.Error()
}

func TestChain(t *testing.T) {
	tests := []struct {
		Name     string
		ToError  func() error
		Expected error
	}{
		{
			"chained error",
			func() error {
				return New("egad")
			},
			&chainedError{
				caller: caller{
					File: "file-1",
					Line: 1,
				},
				current: errors.New("egad"),
				next:    nil,
			},
		},
		{
			"nil chain",
			func() error {
				return Chain(nil)
			},
			nil,
		},
		{
			"nil chain with message",
			func() error {
				return ChainWith(nil, "oh noes")
			},
			nil,
		},
		{
			"chained",
			func() error {
				return New("egad")
			},
			&chainedError{
				caller: caller{
					File: "file-1",
					Line: 1,
				},
				current: errors.New("egad"),
				next:    nil,
			},
		},
		{
			"chained w/ format",
			func() error {
				return Newf("egad %d", 1)
			},
			&chainedError{
				caller: caller{
					File: "file-1",
					Line: 1,
				},
				current: errors.New("egad 1"),
				next:    nil,
			},
		},
		{
			"chaining an unchained",
			func() error {
				return Chain(errors.New("egad"))
			},
			&chainedError{
				caller: caller{
					File: "file-1",
					Line: 1,
				},
				current: errors.New("egad"),
				next:    nil,
			},
		},
		{
			"chaining a chain",
			func() error {
				return Chain(New("egad"))
			},
			&chainedError{
				caller: caller{
					File: "file-2",
					Line: 2,
				},
				next: &chainedError{
					caller: caller{
						File: "file-1",
						Line: 1,
					},
					current: errors.New("egad"),
					next:    nil,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			done := setCallerFunc(newSerialCaller())
			defer done()
			if e := test.ToError(); !isSameError(e, test.Expected) {
				t.Fatalf("expected %v but got %v", test.Expected, e)
			}
		})
	}
}
