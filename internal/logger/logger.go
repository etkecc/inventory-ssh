package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/etkecc/inventory-ssh/internal/tui"
)

const (
	ansiReset = "\033[0m"
	ansiDim   = "\033[2m"
	prefix    = "[inventory-ssh] "
)

var withDebug bool

// Configure the logger package.
func Configure(debug bool) {
	withDebug = debug
}

// Println logs the arguments as an info message.
func Println(args ...any) {
	tui.Info(prefix + sprint(args...))
}

// Debug logs the arguments to stderr if the debug flag is set.
func Debug(args ...any) {
	if !withDebug {
		return
	}
	fmt.Fprintln(os.Stderr, ansiDim+prefix+sprint(args...)+ansiReset)
}

// Fatal logs the arguments as an error and exits.
func Fatal(args ...any) {
	tui.Error(prefix + sprint(args...))
	os.Exit(1)
}

func sprint(args ...any) string {
	return strings.TrimSuffix(fmt.Sprintln(args...), "\n")
}
