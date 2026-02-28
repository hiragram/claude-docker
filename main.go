package main

import (
	"os"

	"github.com/hiragram/agent-workspace/internal/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
