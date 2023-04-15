package main

import "github.com/modernice/opendocs/go/cli"

func main() {
	app := cli.New()

	if err := app.Run(); err != nil {
		app.FatalIfErrorf(err)
	}
}
