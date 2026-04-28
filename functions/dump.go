// Package functions provides utility functions for gotg.
// Includes peer resolution, dump utilities, message helpers, and optional parameter patterns.

package functions

import (
	"fmt"

	"github.com/bytedance/sonic"
)

// Dumpable allows types to control what gets serialized by Dump.
type Dumpable interface {
	DumpValue() any
}

// Dump returns a pretty-printed JSON representation of any value.
// If key is provided, the output is prefixed with [key].
// Types implementing Dumpable will have DumpValue() serialized instead.
func Dump(val any, key ...string) string {
	prefix := ""
	if len(key) > 0 && key[0] != "" {
		prefix = fmt.Sprintf("[%s] ", key[0])
	}
	target := val
	if d, ok := val.(Dumpable); ok {
		target = d.DumpValue()
	}
	data, _ := sonic.MarshalIndent(target, "", "  ")
	return prefix + string(data)
}
