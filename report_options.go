package poop

import (
	"io"
	"os"
)

type ReportOptions struct {
	reporter   func(w io.Writer, err error) error
	terminater func(err error)
}

type ReportOption func(*ReportOptions)

func ExitWithStatus(status int) ReportOption {
	return func(o *ReportOptions) {
		o.terminater = func(err error) {
			os.Exit(status)
		}
	}
}

func Panic() ReportOption {
	return func(o *ReportOptions) {
		o.terminater = func(err error) {
			panic(err)
		}
	}
}

func UsingReporter(reporter func(w io.Writer, err error) error) ReportOption {
	return func(o *ReportOptions) {
		o.reporter = reporter
	}
}
