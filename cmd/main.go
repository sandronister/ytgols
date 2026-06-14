// Command ytgols starts the interactive YouTube downloader.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sandronister/ytgols/internal/app"
	"github.com/sandronister/ytgols/internal/converter"
	"github.com/sandronister/ytgols/internal/downloader"
	"github.com/sandronister/ytgols/internal/form"
)

func main() {
	application := app.New(
		form.New(os.Stdin, os.Stdout),
		downloader.New(converter.NewFFmpeg()),
		os.Stderr,
	)
	_, err := application.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nerro: %v\n", err)
		os.Exit(1)
	}
}
