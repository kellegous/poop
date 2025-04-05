package main

import "github.com/kellegous/poop"

func foo() error {
	return poop.Chain(bar())
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
