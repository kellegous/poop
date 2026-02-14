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

type serialCaller int

func (c serialCaller) frame() Frame {
	return Frame{
		Function: fmt.Sprintf("func-%d", c),
		File:     fmt.Sprintf("file-%d", c),
		Line:     int(c),
	}
}

func newSerialCaller() func() caller {
	i := 0
	return func() caller {
		i++
		return serialCaller(i)
	}
}

func isSameError(a, b error) bool {
	if a == b {
		return true
	} else if a == nil || b == nil {
		return false
	}

	ca, aok := a.(*ChainedError)
	cb, bok := b.(*ChainedError)
	if aok != bok {
		return false
	} else if aok {
		return ca.message == cb.message &&
			isSameError(ca.next, cb.next) &&
			ca.caller.frame() == cb.caller.frame()
	}

	return a.Error() == b.Error()
}

func describe(err error) string {
	if err == nil {
		return "nil"
	}

	c, ok := err.(*ChainedError)
	if !ok {
		return fmt.Sprintf("error(%s)", err.Error())
	}
	f := c.caller.frame()
	return fmt.Sprintf(
		"chained(current: %s, caller: [%s,%s,%d], next: %s)",
		c.message,
		f.Function,
		f.File,
		f.Line,
		describe(c.next))
}

type mockCaller Frame

func (c mockCaller) frame() Frame {
	return Frame(c)
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
			&ChainedError{
				caller: mockCaller(Frame{
					Function: "func-1",
					File:     "file-1",
					Line:     1,
				}),
				message: "egad",
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
			&ChainedError{
				caller: mockCaller(Frame{
					Function: "func-1",
					File:     "file-1",
					Line:     1,
				}),
				message: "egad",
				next:    nil,
			},
		},
		{
			"chained w/ format",
			func() error {
				return Newf("egad %d", 1)
			},
			&ChainedError{
				caller: mockCaller(Frame{
					Function: "func-1",
					File:     "file-1",
					Line:     1,
				}),
				message: "egad 1",
				next:    nil,
			},
		},
		{
			"chaining an unchained",
			func() error {
				return Chain(errors.New("egad"))
			},
			&ChainedError{
				caller: mockCaller(Frame{
					Function: "func-1",
					File:     "file-1",
					Line:     1,
				}),
				message: "",
				next:    errors.New("egad"),
			},
		},
		{
			"chaining a chain",
			func() error {
				return Chain(New("egad"))
			},
			&ChainedError{
				caller: mockCaller(Frame{
					Function: "func-2",
					File:     "file-2",
					Line:     2,
				}),
				next: &ChainedError{
					caller: mockCaller(Frame{
						Function: "func-1",
						File:     "file-1",
						Line:     1,
					}),
					message: "egad",
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
				t.Fatalf("expected:\n%s\nbut got\n%s", describe(test.Expected), describe(e))
			}
		})
	}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

var theError = &customError{msg: "theError"}

func TestIs(t *testing.T) {
	tests := []struct {
		Name  string
		Error error
		Check func(t *testing.T, err error)
	}{
		{
			"single",
			Chain(theError),
			func(t *testing.T, err error) {
				if !errors.Is(err, theError) {
					t.Errorf("expected %v to be %v", err, theError)
				}
			},
		},
		{
			"single_as",
			Chain(theError),
			func(t *testing.T, err error) {
				var cerr *customError
				if !errors.As(err, &cerr) || cerr.msg != "theError" {
					t.Errorf("expected %v to be %v", err, theError)
				}
			},
		},
		{
			"double",
			Chain(Chain(theError)),
			func(t *testing.T, err error) {
				if !errors.Is(err, theError) {
					t.Errorf("expected %v to be %v", err, theError)
				}
			},
		},
		{
			"double_as",
			Chain(Chain(theError)),
			func(t *testing.T, err error) {
				var cerr *customError
				if !errors.As(err, &cerr) || cerr.msg != "theError" {
					t.Errorf("expected %v to be %v", err, theError)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Check(t, test.Error)
		})
	}
}
