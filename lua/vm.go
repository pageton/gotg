package lua

import (
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

const gotgModuleName = "gotg"

type VM struct {
	L  *lua.LState
	mu sync.Mutex
}

func NewVM() *VM {
	L := lua.NewState()
	vm := &VM{L: L}
	vm.registerTypes()
	return vm
}

func (vm *VM) registerTypes() {
	registerClientType(vm.L)
	registerUpdateType(vm.L)
	vm.L.PreloadModule(gotgModuleName, vm.loader)
}

func (vm *VM) loader(L *lua.LState) int {
	mod := L.NewTable()

	L.SetField(mod, "new_client", L.NewFunction(luaNewClient))

	L.Push(mod)
	return 1
}

func (vm *VM) InjectSelf() {
	ud := vm.L.NewUserData()
	ud.Value = vm
	reg := vm.L.Get(lua.RegistryIndex)
	if tbl, ok := reg.(*lua.LTable); ok {
		tbl.RawSetString("__gotg_vm", ud)
	}
}

func (vm *VM) DoFile(path string) error {
	return vm.L.DoFile(path)
}

func (vm *VM) DoString(code string) error {
	return vm.L.DoString(code)
}

func (vm *VM) Close() {
	vm.L.Close()
}

func (vm *VM) CallLuaFunc(fn *lua.LFunction, args ...lua.LValue) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	err := vm.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, args...)
	if err != nil {
		return fmt.Errorf("lua handler error: %w", err)
	}
	return nil
}
