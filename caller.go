package poop

import "runtime"

type caller struct {
	File string
	Line int
}

var callerFunc = defaultCallerFunc

func defaultCallerFunc() caller {
	_, file, line, _ := runtime.Caller(2)
	return caller{
		File: file,
		Line: line,
	}
}

func setCallerFunc(f func() caller) func() {
	orig := callerFunc
	callerFunc = f
	return func() {
		callerFunc = orig
	}
}
