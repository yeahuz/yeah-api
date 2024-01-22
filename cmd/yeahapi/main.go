package main

import (
	"fmt"
	"os"

	"github.com/yeahuz/yeah-api/serverutil/backend"
)

func main() {
	if err := backend.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
