package main

import (
	"fmt"
	"os"

	"github.com/clambin/xcoder/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
