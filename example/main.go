package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/kellegous/poop"
)

type UseRoot string

const (
	RootIsChainError    UseRoot = "is-chain"
	RootIsNotChainError UseRoot = "is-not-chain"
	RootIsCompressable  UseRoot = "is-not-chain-uncompressable"
	RootHasUnwrap       UseRoot = "is-not-chain-has-unwrap"
)

var (
	useRoot UseRoot = RootIsChainError
)

func init() {
	flag.Var(&useRoot, "use-root", "the use root value (default: root-is-chain-error)")
	flag.Parse()
}

func (u *UseRoot) Set(v string) error {
	*u = UseRoot(v)
	return u.validate()
}

func (u UseRoot) validate() error {
	switch u {
	case RootIsChainError, RootIsNotChainError, RootIsCompressable, RootHasUnwrap:
		return nil
	default:
		return fmt.Errorf("invalid use-root value: %s", u)
	}
}

func (u UseRoot) String() string {
	return string(u)
}

func foo() error {
	return poop.ChainWith(bar(), "this is a really long message that is meant to test the table formatter's ability to truncate text")
}

func bar() error {
	return poop.Chain(baz(5))
}

func baz(n int) error {
	if n == 0 {
		switch useRoot {
		case RootIsChainError:
			return poop.Newf("n is zero")
		case RootIsNotChainError:
			return errors.New("n is zero")
		case RootIsCompressable:
			return poop.ChainWith(errors.New("n is zero"), "zero has been reached")
		case RootHasUnwrap:
			return fmt.Errorf("zero has been reached: %w", errors.New("n is zero"))
		}
	}
	return poop.Chain(baz(n - 1))
}

func main() {
	if err := foo(); err != nil {
		poop.HitFan(err)
	}
}
