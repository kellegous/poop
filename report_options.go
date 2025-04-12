package poop

import (
	"io"
	"os"
)

var defaultReportOptions = ReportOptions{
	reporter:   DefaultReporter,
	terminater: exitTerminator(1),
}

type ReportOptions struct {
	reporter   Reporter
	terminater func(err error)
}

type ReportOption func(*ReportOptions)

type Reporter func(w io.Writer, err error) error

func exitTerminator(status int) func(err error) {
	return func(err error) {
		os.Exit(status)
	}
}

func ExitWithStatus(status int) ReportOption {
	return func(o *ReportOptions) {
		o.terminater = exitTerminator(1)
	}
}

func Panic() ReportOption {
	return func(o *ReportOptions) {
		o.terminater = func(err error) {
			panic(err)
		}
	}
}

func UsingReporter(reporter Reporter) ReportOption {
	return func(o *ReportOptions) {
		o.reporter = reporter
	}
}
