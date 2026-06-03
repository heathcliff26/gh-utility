package main

import (
	"os"
)

func main() {
	cmd := NewRoot()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
