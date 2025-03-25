package poop

import (
	"errors"
	"os"
)

type SHTF interface {
	HitFan(err error)
}

type shtf struct {
	opts ReportOptions
}

func (s *shtf) HitFan(err error) {
	hitFan(err, &s.opts, callerFunc())
}

func BeReady(opts ...ReportOption) SHTF {
	return &shtf{
		opts: ReportOptions{
			reporter: DefaultReporter,
			terminater: func(err error) {
				os.Exit(1)
			},
		},
	}
}

func hitFan(err error, opts *ReportOptions, caller Caller) {
	err = newChainedError(err, errors.New("the 💩 hath hit the fan"), caller)
	if err := opts.reporter(os.Stderr, err); err != nil {
		panic(err)
	}

	opts.terminater(err)
}

func HitFan(err error, opts ...ReportOption) {
	o := ReportOptions{
		reporter: DefaultReporter,
		terminater: func(err error) {
			os.Exit(1)
		},
	}

	for _, opt := range opts {
		opt(&o)
	}

	hitFan(err, &o, callerFunc())
}
