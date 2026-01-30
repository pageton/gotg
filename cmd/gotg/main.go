package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		requireScript(os.Args[2:])
		cmdRun(os.Args[2])
	case "dev":
		requireScript(os.Args[2:])
		cmdDev(os.Args[2])
	case "build":
		cmdBuild(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func requireScript(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "error: script path required\n\n")
		printUsage()
		os.Exit(1)
	}
	if _, err := os.Stat(args[0]); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: file not found: %s\n", args[0])
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `gotg - Telegram bot framework for Lua

usage:
  gotg run   <script.lua>              run a lua script
  gotg dev   <script.lua>              run with auto-restart on file changes
  gotg build <script.lua> [-o output]  compile into a standalone binary
`)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
