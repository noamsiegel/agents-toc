// agents-toc keeps a one-line lazy-load index inside AGENTS.md in sync with
// the project's skill and knowledge markdown sources.
//
// See README.md and `agents-toc --help`.
package main

import (
	"fmt"
	"os"

	"github.com/noamsiegel/agents-toc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
