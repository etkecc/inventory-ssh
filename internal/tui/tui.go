// Package tui provides terminal output helpers with consistent color formatting.
package tui

import (
	"fmt"
	"os"
)

// ANSI escape codes for terminal colors.
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

// Success prints a labeled success value in bold green.
func Success(label, value string) {
	fmt.Println(ansiBold + ansiGreen + label + ":" + ansiReset + " " + value)
}

// Info prints an informational message in cyan.
func Info(msg string) {
	fmt.Println(ansiCyan + msg + ansiReset)
}

// Warning prints a warning to stderr in bold yellow.
func Warning(msg string) {
	fmt.Fprintln(os.Stderr, ansiBold+ansiYellow+"WARNING:"+ansiReset+" "+msg)
}

// Error prints an error to stderr in bold red.
func Error(msg string) {
	fmt.Fprintln(os.Stderr, ansiBold+ansiRed+"ERROR:"+ansiReset+" "+msg)
}

// Errorf prints a formatted error to stderr in bold red.
func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}
