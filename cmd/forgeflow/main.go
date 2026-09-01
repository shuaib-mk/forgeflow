package main

import (
	"fmt"
	"os"

	"github.com/forgeflow/forgeflow/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "forgeflow:", err)
		os.Exit(1)
	}
}
