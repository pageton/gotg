package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func cmdDev(script string) {
	absScript, err := filepath.Abs(script)
	if err != nil {
		fatal("cannot resolve path: %s", err)
	}

	watchDir := filepath.Dir(absScript)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		fmt.Fprintf(os.Stderr, "[dev] starting %s\n", filepath.Base(absScript))
		proc := startProcess(absScript)

		exitCh := make(chan error, 1)
		if proc != nil {
			go func() { exitCh <- proc.Wait() }()
		}

		changeCh := pollChanges(watchDir, 500*time.Millisecond)

		select {
		case <-sigCh:
			killProcess(proc)
			fmt.Fprintf(os.Stderr, "\n[dev] stopped\n")
			os.Exit(0)

		case <-changeCh:
			killProcess(proc)
			fmt.Fprintf(os.Stderr, "[dev] change detected, restarting...\n")

		case err := <-exitCh:
			if err != nil {
				fmt.Fprintf(os.Stderr, "[dev] process exited: %s\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "[dev] process exited\n")
			}
			fmt.Fprintf(os.Stderr, "[dev] waiting for changes...\n")
			select {
			case <-pollChanges(watchDir, 500*time.Millisecond):
				fmt.Fprintf(os.Stderr, "[dev] change detected, restarting...\n")
			case <-sigCh:
				fmt.Fprintf(os.Stderr, "\n[dev] stopped\n")
				os.Exit(0)
			}
		}
	}
}

func startProcess(script string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "run", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[dev] failed to start: %s\n", err)
		return nil
	}
	return cmd
}

func killProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}

func pollChanges(dir string, interval time.Duration) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		baseline := scanModTimes(dir)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			current := scanModTimes(dir)
			if modTimesChanged(baseline, current) {
				ch <- struct{}{}
				return
			}
		}
	}()
	return ch
}

func scanModTimes(dir string) map[string]time.Time {
	m := make(map[string]time.Time)
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".lua") {
			m[path] = info.ModTime()
		}
		return nil
	})
	return m
}

func modTimesChanged(old, current map[string]time.Time) bool {
	if len(old) != len(current) {
		return true
	}
	for path, t := range current {
		if prev, ok := old[path]; !ok || !prev.Equal(t) {
			return true
		}
	}
	return false
}
