package log

import (
	"path/filepath"
	"runtime"
)

type callerInfo struct {
	File string
	Line int
	Func string
}

func resolveCaller(skip int) callerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return callerInfo{File: "???", Line: 0}
	}
	ci := callerInfo{
		File: filepath.Base(file),
		Line: line,
	}
	if fn := runtime.FuncForPC(pc); fn != nil {
		ci.Func = filepath.Base(fn.Name())
	}
	return ci
}
