package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dayflower/export-ua-history/internal/app"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, time.Now()); err != nil {
		printError(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, now time.Time) error {
	runner := app.NewRunner()
	return runner.Run(args, stdout, stderr, now)
}

func printError(stderr io.Writer, err error) {
	message := err.Error()
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	_, _ = fmt.Fprint(stderr, message)
}
