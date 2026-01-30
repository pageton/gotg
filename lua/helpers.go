package lua

import (
	lua "github.com/yuin/gopher-lua"
)

func getStringField(tbl *lua.LTable, key string) string {
	v := tbl.RawGetString(key)
	if s, ok := v.(lua.LString); ok {
		return string(s)
	}
	return ""
}

func getNumberField(tbl *lua.LTable, key string) float64 {
	v := tbl.RawGetString(key)
	if n, ok := v.(lua.LNumber); ok {
		return float64(n)
	}
	return 0
}

func getBoolField(tbl *lua.LTable, key string) bool {
	v := tbl.RawGetString(key)
	if b, ok := v.(lua.LBool); ok {
		return bool(b)
	}
	return false
}

func getOptBoolField(tbl *lua.LTable, key string) *bool {
	v := tbl.RawGetString(key)
	if b, ok := v.(lua.LBool); ok {
		val := bool(b)
		return &val
	}
	return nil
}

func getTableField(tbl *lua.LTable, key string) *lua.LTable {
	v := tbl.RawGetString(key)
	if t, ok := v.(*lua.LTable); ok {
		return t
	}
	return nil
}
