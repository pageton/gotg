package lua

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	lua "github.com/yuin/gopher-lua"
)

var (
	inputPeerClassType = reflect.TypeFor[tg.InputPeerClass]()
	inputUserClassType = reflect.TypeFor[tg.InputUserClass]()
)

func updateRawCall(L *lua.LState) int {
	u := checkUpdate(L)
	methodName := L.CheckString(2)
	paramsTbl := L.OptTable(3, nil)
	return doRawCall(L, u.Ctx, methodName, paramsTbl)
}

func doRawCall(L *lua.LState, ctx *adapter.Context, methodName string, paramsTbl *lua.LTable) int {
	goName := tlToGoMethodName(methodName)
	raw := reflect.ValueOf(ctx.Raw)
	method := raw.MethodByName(goName)
	if !method.IsValid() {
		L.ArgError(2, "unknown method: "+methodName+" (tried "+goName+")")
		return 0
	}

	methodType := method.Type()
	numIn := methodType.NumIn()

	var results []reflect.Value

	switch numIn {
	case 1:
		results = method.Call([]reflect.Value{
			reflect.ValueOf(ctx.Context),
		})
	case 2:
		reqType := methodType.In(1)
		var reqVal reflect.Value

		switch reqType.Kind() {
		case reflect.Ptr:
			reqVal = reflect.New(reqType.Elem())
			if paramsTbl != nil {
				populateStruct(reqVal.Elem(), paramsTbl, ctx)
			}
		case reflect.Interface:
			var lv lua.LValue
			if paramsTbl != nil {
				lv = luaTableFirstValue(paramsTbl)
			}
			if lv == nil || lv == lua.LNil {
				lv = L.Get(3)
			}
			resolved, ok := resolveInterfaceField(reqType, lv, ctx)
			if !ok {
				L.ArgError(3, "cannot resolve interface param for: "+methodName)
				return 0
			}
			reqVal = reflect.ValueOf(resolved)
		case reflect.Struct:
			reqVal = reflect.New(reqType).Elem()
			if paramsTbl != nil {
				populateStruct(reqVal, paramsTbl, ctx)
			}
		default:
			reqVal = reflect.New(reqType).Elem()
			if paramsTbl != nil {
				setField(reqVal, reqType, paramsTbl, ctx)
			}
		}

		results = method.Call([]reflect.Value{
			reflect.ValueOf(ctx.Context),
			reqVal,
		})
	default:
		L.ArgError(2, fmt.Sprintf("unsupported method signature (%d params) for: %s", numIn, methodName))
		return 0
	}

	if methodType.NumOut() == 2 {
		errVal := results[1]
		if !errVal.IsNil() {
			L.Push(lua.LNil)
			L.Push(lua.LString(errVal.Interface().(error).Error()))
			return 2
		}
	}

	resp := results[0]
	if !resp.IsValid() || (resp.Kind() == reflect.Pointer && resp.IsNil()) {
		L.Push(lua.LTrue)
		return 1
	}

	L.Push(goToLua(L, resp))
	return 1
}

func populateStruct(v reflect.Value, tbl *lua.LTable, ctx *adapter.Context) {
	t := v.Type()
	fieldMap := buildFieldMap(t)

	tbl.ForEach(func(key, val lua.LValue) {
		ks, ok := key.(lua.LString)
		if !ok {
			return
		}
		idx, ok := fieldMap[string(ks)]
		if !ok {
			return
		}
		field := v.Field(idx)
		fieldType := t.Field(idx)
		setField(field, fieldType.Type, val, ctx)
	})
}

func buildFieldMap(t reflect.Type) map[string]int {
	m := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := jsonFieldName(f)
		if name == "-" {
			continue
		}
		m[name] = i
		lower := strings.ToLower(f.Name)
		if lower != name {
			m[lower] = i
		}
	}
	return m
}

func setField(field reflect.Value, fieldType reflect.Type, lv lua.LValue, ctx *adapter.Context) {
	if resolved, ok := resolveInterfaceField(fieldType, lv, ctx); ok {
		field.Set(reflect.ValueOf(resolved))
		return
	}

	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
		ptr := reflect.New(fieldType)
		setField(ptr.Elem(), fieldType, lv, ctx)
		field.Set(ptr)
		return
	}

	switch fieldType.Kind() {
	case reflect.Bool:
		if b, ok := lv.(lua.LBool); ok {
			field.SetBool(bool(b))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, ok := lv.(lua.LNumber); ok {
			field.SetInt(int64(n))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, ok := lv.(lua.LNumber); ok {
			field.SetUint(uint64(n))
		}
	case reflect.Float32, reflect.Float64:
		if n, ok := lv.(lua.LNumber); ok {
			field.SetFloat(float64(n))
		}
	case reflect.String:
		if s, ok := lv.(lua.LString); ok {
			field.SetString(string(s))
		}
	case reflect.Slice:
		if tbl, ok := lv.(*lua.LTable); ok {
			elemType := fieldType.Elem()
			n := tbl.MaxN()
			slice := reflect.MakeSlice(fieldType, 0, n)
			for i := 1; i <= n; i++ {
				elem := reflect.New(elemType).Elem()
				setField(elem, elemType, tbl.RawGetInt(i), ctx)
				slice = reflect.Append(slice, elem)
			}
			field.Set(slice)
		}
	case reflect.Struct:
		if tbl, ok := lv.(*lua.LTable); ok {
			populateStruct(field, tbl, ctx)
		}
	case reflect.Interface:
		if resolved, ok := resolveInterfaceField(fieldType, lv, ctx); ok {
			field.Set(reflect.ValueOf(resolved))
		}
	}
}

func resolveInterfaceField(fieldType reflect.Type, lv lua.LValue, ctx *adapter.Context) (any, bool) {
	if fieldType.Kind() != reflect.Interface {
		return nil, false
	}

	switch {
	case fieldType.Implements(inputPeerClassType) || fieldType == inputPeerClassType:
		if num, ok := lv.(lua.LNumber); ok {
			peer := ctx.PeerStorage.GetInputPeerByID(int64(num))
			if peer != nil {
				return peer, true
			}
		}

	case fieldType.Implements(inputUserClassType) || fieldType == inputUserClassType:
		switch v := lv.(type) {
		case lua.LNumber:
			return &tg.InputUser{UserID: int64(v)}, true
		case *lua.LTable:
			userID := int64(getNumberField(v, "user_id"))
			accessHash := int64(getNumberField(v, "access_hash"))
			return &tg.InputUser{UserID: userID, AccessHash: accessHash}, true
		}
	}

	return nil, false
}

func luaTableFirstValue(tbl *lua.LTable) lua.LValue {
	if tbl == nil {
		return lua.LNil
	}
	var first lua.LValue
	tbl.ForEach(func(_, v lua.LValue) {
		if first == nil {
			first = v
		}
	})
	if first == nil {
		return lua.LNil
	}
	return first
}

func tlToGoMethodName(name string) string {
	parts := strings.Split(name, ".")
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}
