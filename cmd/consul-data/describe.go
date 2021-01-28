package main

import (
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
)

type describeCommand struct {
	ui         cli.Ui
	configPath string
	randSeed   int64

	flags *flag.FlagSet
	help  string
}

func newDescribeCommand(ui cli.Ui) cli.Command {
	c := &describeCommand{
		ui: ui,
	}

	flags := flag.NewFlagSet("", flag.ContinueOnError)

	c.flags = flags
	c.help = genUsage(`Usage: consul-data describe [OPTIONS] <data path>
	
	Describe contents of the randomly generated data file.`, c.flags)

	return c
}

func (c *describeCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to parse command line arguments: %v", err))
		return 1
	}

	args = c.flags.Args()
	if len(args) < 1 {
		c.ui.Error(fmt.Sprintf("Must supply the path to the data as a positional argument"))
		return 1
	}

	data, err := loadData(args[0])
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	svcCount := 0
	for _, node := range data.Catalog {
		svcCount += len(node.Services)
	}

	c.ui.Info(fmt.Sprintf("Keys:     %d", len(data.KV)))
	c.ui.Info(fmt.Sprintf("Nodes:    %d", len(data.Catalog)))
	c.ui.Info(fmt.Sprintf("services: %d", svcCount))

	return 0
}

func (c *describeCommand) Synopsis() string {
	return "Describe generated Consul data for Consul"
}

func (c *describeCommand) Help() string {
	return c.help
}
