package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("consul-data-gen", "0.0.1")
	c.Args = os.Args[1:]

	ui := &cli.BasicUi{Reader: os.Stdin, Writer: os.Stdout, ErrorWriter: os.Stderr}

	c.Commands = map[string]cli.CommandFactory{
		"kv": func() (cli.Command, error) { return newKVCommand(ui), nil },
		// "catalog": catalogCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

// func catalogCommandFactory() (cli.Command, error) {

// }
