package main

import (
	"log/slog"
	"os"
)

func main() {
	cmd := NewRoot()
	err := cmd.Execute()
	if err != nil {
		slog.Error("Command exited with error", "err", err)
		os.Exit(1)
	}
}
