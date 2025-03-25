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

// BeReady creates a new SHTF instance with the given options. This allows
// for the creation of customized reporting behavior in a specific area of code,
// without affecting the global default behavior.
func BeReady(opts ...ReportOption) SHTF {
	return &shtf{
		opts: defaultReportOptions,
	}
}

// Configure modifies the default reporting options changing the global default
// behavior of HitFan. This is intended for 1-time configuration in main.
func Configure(opts ...ReportOption) {
	for _, opt := range opts {
		opt(&defaultReportOptions)
	}
}

func hitFan(err error, opts *ReportOptions, caller caller) {
	err = newChainedError(err, errors.New("the 💩 hath hit the fan"), caller)
	if err := opts.reporter(os.Stderr, err); err != nil {
		panic(err)
	}

	opts.terminater(err)
}

// HitFan reports the given error and then terminates the program using the
// report options to control how the error is reported and the mechanism by
// which the program is terminated.
func HitFan(err error, opts ...ReportOption) {
	o := defaultReportOptions
	for _, opt := range opts {
		opt(&o)
	}
	hitFan(err, &o, callerFunc())
}
