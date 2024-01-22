package main

import (
	"fmt"
	"os"

	"github.com/yeahuz/yeah-api/serverutil/frontend"
)

func main() {
	if err := frontend.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
