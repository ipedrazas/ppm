// Command ppm is a CLI for the PM/Product-Owner agent's memory system.
// The on-disk format is defined in plans/memory-format.md.
package main

import (
	"os"

	"github.com/ipedrazas/ppm/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
