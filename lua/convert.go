package lua

import (
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func goToLua(L *lua.LState, v reflect.Value) lua.LValue {
	if !v.IsValid() {
		return lua.LNil
	}

	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return lua.LNil
		}
		return goToLua(L, v.Elem())
	}

	switch v.Kind() {
	case reflect.Bool:
		return lua.LBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(v.Uint())
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(v.Float())
	case reflect.String:
		return lua.LString(v.String())
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return lua.LString(string(v.Bytes()))
		}
		tbl := L.NewTable()
		for i := 0; i < v.Len(); i++ {
			tbl.RawSetInt(i+1, goToLua(L, v.Index(i)))
		}
		return tbl
	case reflect.Map:
		tbl := L.NewTable()
		for _, key := range v.MapKeys() {
			tbl.RawSetString(key.String(), goToLua(L, v.MapIndex(key)))
		}
		return tbl
	case reflect.Struct:
		return structToLua(L, v)
	default:
		return lua.LNil
	}
}

func structToLua(L *lua.LState, v reflect.Value) *lua.LTable {
	tbl := L.NewTable()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fv := v.Field(i)
		if fv.IsZero() {
			continue
		}
		key := jsonFieldName(field)
		if key == "-" {
			continue
		}
		tbl.RawSetString(key, goToLua(L, fv))
	}
	return tbl
}

func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return toSnakeCase(f.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return toSnakeCase(f.Name)
	}
	return name
}

func toSnakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteByte(byte(r) + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func luaToGo(lv lua.LValue) any {
	switch v := lv.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		n := float64(v)
		if n == float64(int64(n)) {
			return int64(n)
		}
		return n
	case lua.LString:
		return string(v)
	case *lua.LTable:
		return luaTableToMap(v)
	case *lua.LNilType:
		return nil
	default:
		return nil
	}
}

func luaTableToMap(tbl *lua.LTable) any {
	isArray := true
	maxN := tbl.MaxN()
	if maxN == 0 {
		isArray = false
	}

	if isArray {
		arr := make([]any, 0, maxN)
		for i := 1; i <= maxN; i++ {
			arr = append(arr, luaToGo(tbl.RawGetInt(i)))
		}
		return arr
	}

	m := make(map[string]any)
	tbl.ForEach(func(k, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			m[string(ks)] = luaToGo(v)
		}
	})
	return m
}
