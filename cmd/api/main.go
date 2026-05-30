package main

import (
	"fmt"
	"os"

	"github.com/radius/radius-backend/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
