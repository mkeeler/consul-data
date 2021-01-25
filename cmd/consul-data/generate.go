package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/mitchellh/cli"
	"github.com/mkeeler/consul-data/generate"
)

type generateCommand struct {
	ui         cli.Ui
	configPath string
	randSeed   int64

	flags *flag.FlagSet
	help  string
}

func newGenerateCommand(ui cli.Ui) cli.Command {
	c := &generateCommand{
		ui: ui,
	}

	flags := flag.NewFlagSet("", flag.ContinueOnError)

	flags.Int64Var(&c.randSeed, "seed", 0, "Value to use to seed the pseudo-random number generator with instead of the current time")
	flags.StringVar(&c.configPath, "config", "", "Path to the configuration to use for generating data")

	c.flags = flags
	c.help = genUsage(`Usage: consul-data generate [OPTIONS] [output path]
	
	Generate random data for consul.
	
	By default the generated output is sent to the console but
	an optional output path may be used to cause it to be written
	to a file`, c.flags)

	return c
}

func (c *generateCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to parse command line arguments: %v", err))
		return 1
	}

	args = c.flags.Args()

	if c.randSeed == 0 {
		c.randSeed = time.Now().UnixNano()
	}
	rand.Seed(c.randSeed)

	conf := generate.DefaultConfig()

	if c.configPath != "" {
		var err error
		conf, err = generate.ParseConfig(c.configPath)
		if err != nil {
			c.ui.Error(err.Error())
			return 1
		}
	}

	data, err := generate.GenerateAll(conf)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to generate Consul data: %v", err))
		return 1
	}

	serialized, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to serialize Consul data: %v", err))
		return 1
	}

	if len(args) > 0 {
		if err := ioutil.WriteFile(args[0], serialized, 0644); err != nil {
			c.ui.Error(fmt.Sprintf("Failed to write serialized Consul data to %q: %v", args[0], err))
			return 1
		}
		c.ui.Info(fmt.Sprintf("Consul data written to %s", args[0]))
	} else {
		c.ui.Info(string(serialized))
	}

	return 0
}

func (c *generateCommand) Synopsis() string {
	return "Generate Consul data for Consul"
}

func (c *generateCommand) Help() string {
	return c.help
}
