package log

import "time"

type Record struct {
	Level   Level
	Message string
	Time    time.Time
	Module  string
	File    string
	Line    int
	Func    string
	Fields  []Field
}

type Field struct {
	Key   string
	Value any
}
