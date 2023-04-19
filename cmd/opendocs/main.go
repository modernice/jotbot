package main

import (
	"log"
	"os"

	"github.com/modernice/opendocs/cli"
)

func main() {
	log.SetFlags(0)

	app := cli.New()

	if err := app.Run(); err != nil {
		app.FatalIfErrorf(err)
	}

	os.Exit(0)
}
