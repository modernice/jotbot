package main

import (
	"flag"
	"log"
	"os"

	opendocs "github.com/modernice/opendocs/go"
)

var (
	filepath      string
	typeName      string
	documentation string
)

func init() {
	log.SetFlags(0)

	flag.StringVar(&filepath, "file", "", "File to patch")
	flag.StringVar(&filepath, "f", "", "File to patch")
	flag.StringVar(&typeName, "type", "", "Type to add documentation to")
	flag.StringVar(&typeName, "t", "", "Type to add documentation to")
	flag.StringVar(&documentation, "doc", "", "Documentation to add")
	flag.StringVar(&documentation, "d", "", "Documentation to add")
	flag.Parse()
}

func main() {
	if filepath == "" || typeName == "" || documentation == "" {
		flag.Usage()
		return
	}

	p, err := opendocs.NewPatcher(filepath)
	if err != nil {
		log.Panic(err)
	}

	if err := p.Patch(typeName, documentation); err != nil {
		log.Panic(err)
	}

	os.Exit(0)
}
