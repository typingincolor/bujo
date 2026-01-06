package main

import (
	"os"

	"github.com/typingincolor/bujo/cmd/bujo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
