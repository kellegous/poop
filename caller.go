package poop

import "runtime"

type Caller struct {
	File string
	Line int
}

var callerFunc = defaultCallerFunc

func defaultCallerFunc() Caller {
	_, file, line, _ := runtime.Caller(2)
	return Caller{
		File: file,
		Line: line,
	}
}

func setCallerFunc(f func() Caller) func() {
	orig := callerFunc
	callerFunc = f
	return func() {
		callerFunc = orig
	}
}
