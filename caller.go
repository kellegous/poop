package poop

import "runtime"

type caller interface {
	frame() frame
}

type runtimeCaller uintptr

func (c runtimeCaller) frame() frame {
	pcs := []uintptr{uintptr(c)}
	f, _ := runtime.CallersFrames(pcs).Next()
	return frame{
		function: f.Function,
		file:     f.File,
		line:     f.Line,
	}
}

type frame struct {
	function string
	file     string
	line     int
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
