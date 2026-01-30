package main

import (
	gotglua "github.com/pageton/gotg/lua"
)

func cmdRun(script string) {
	vm := gotglua.NewVM()
	defer vm.Close()

	vm.InjectSelf()

	if err := vm.DoFile(script); err != nil {
		fatal("lua: %s", err)
	}
}
