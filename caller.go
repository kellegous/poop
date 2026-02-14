package poop

import "runtime"

type caller interface {
	frame() Frame
}

type runtimeCaller uintptr

func (c runtimeCaller) frame() Frame {
	pcs := []uintptr{uintptr(c)}
	f, _ := runtime.CallersFrames(pcs).Next()
	return Frame{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

type Frame struct {
	Function string
	File     string
	Line     int
}

var callerFunc = defaultCallerFunc

func defaultCallerFunc() caller {
	pcs := make([]uintptr, 1)
	n := runtime.Callers(3, pcs)
	if n == 0 {
		panic("no caller found")
	}
	return runtimeCaller(pcs[0])
}

func setCallerFunc(f func() caller) func() {
	orig := callerFunc
	callerFunc = f
	return func() {
		callerFunc = orig
	}
}
