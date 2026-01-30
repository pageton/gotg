package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

func cmdBuild(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "error: script path required\n\n")
		printUsage()
		os.Exit(1)
	}

	script := args[0]
	if _, err := os.Stat(script); os.IsNotExist(err) {
		fatal("file not found: %s", script)
	}

	output := strings.TrimSuffix(filepath.Base(script), ".lua")
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	for i := 1; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			output = args[i+1]
			i++
		}
	}

	absOutput, err := filepath.Abs(output)
	if err != nil {
		fatal("cannot resolve output path: %s", err)
	}

	scriptData, err := os.ReadFile(script)
	if err != nil {
		fatal("cannot read script: %s", err)
	}

	tmpDir, err := os.MkdirTemp("", "gotg-build-*")
	if err != nil {
		fatal("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "script.lua"), scriptData, 0644); err != nil {
		fatal("cannot write script: %s", err)
	}

	mainGo := `package main

import (
	_ "embed"
	"fmt"
	"os"

	gotglua "github.com/pageton/gotg/lua"
)

//go:embed script.lua
var luaScript string

func main() {
	vm := gotglua.NewVM()
	defer vm.Close()
	vm.InjectSelf()
	if err := vm.DoString(luaScript); err != nil {
		fmt.Fprintf(os.Stderr, "lua error: %s\n", err)
		os.Exit(1)
	}
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		fatal("cannot write main.go: %s", err)
	}

	modPath, modVersion, localDir := gotgModuleInfo()
	goMod := fmt.Sprintf("module gotg-build\n\ngo 1.24.0\n\nrequire %s %s\n", modPath, modVersion)
	if localDir != "" {
		goMod += fmt.Sprintf("\nreplace %s => %s\n", modPath, localDir)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		fatal("cannot write go.mod: %s", err)
	}

	fmt.Fprintf(os.Stderr, "building %s -> %s\n", filepath.Base(script), output)

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmpDir
	tidy.Stderr = os.Stderr
	if err := tidy.Run(); err != nil {
		fatal("go mod tidy failed: %s", err)
	}

	build := exec.Command("go", "build", "-o", absOutput, ".")
	build.Dir = tmpDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fatal("go build failed: %s", err)
	}

	fmt.Fprintf(os.Stderr, "built: %s\n", output)
}

func gotgModuleInfo() (modPath, version, localDir string) {
	modPath = "github.com/pageton/gotg"

	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return modPath, info.Main.Version, ""
	}

	dir, _ := os.Getwd()
	for dir != "/" && dir != "." {
		gomod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(gomod); err == nil {
			if strings.Contains(string(data), "module "+modPath) {
				return modPath, "v0.0.0", dir
			}
		}
		dir = filepath.Dir(dir)
	}

	return modPath, "latest", ""
}
