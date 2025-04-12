package main

import "github.com/kellegous/poop"

func foo() error {
	return poop.ChainWith(bar(), "this is a really long message that is meant to test the table formatter's ability to truncate text")
}

func bar() error {
	return poop.Chain(baz(5))
}

func baz(n int) error {
	if n == 0 {
		return poop.Newf("n is zero")
	}
	return poop.Chain(baz(n - 1))
}

func main() {
	if err := foo(); err != nil {
		poop.HitFan(err)
	}
}
