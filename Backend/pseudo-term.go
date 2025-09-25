//go:build ptyrepl
// +build ptyrepl

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Simple REPL that runs each entered line via powershell.exe using os/exec.
// Keep the build tag so you can run with: go run -tags ptyrepl pseudo-term.go
func main() {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := in.ReadString('\n')
		if err != nil {
			return
		}
		cmdStr := strings.TrimSpace(line)
		if cmdStr == "" {
			continue
		}
		if cmdStr == "exit" {
			return
		}

		out, err := runCommandExec(cmdStr)
		if len(out) > 0 {
			fmt.Print(out)
		}
		if err != nil && len(out) == 0 {
			fmt.Fprintf(os.Stderr, "(error) %v\n", err)
		}
	}
}

// runCommandExec runs the provided command string using powershell -Command and
// returns combined stdout+stderr as a string. A timeout is used to avoid hangs.
func runCommandExec(cmdStr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", cmdStr)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
