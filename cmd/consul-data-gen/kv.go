package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/mitchellh/cli"

	"github.com/mkeeler/consul-data-gen/generate/kv"
)

const (
	keyTypePetName = "pet-name"

	valueTypeRandomB64 = "b64-random"
)

type kvCommand struct {
	ui     cli.Ui
	config kv.Config

	keyType   stringChoiceValue
	valueType stringChoiceValue

	numEntries int

	randValueMinSize int
	randValueMaxSize int

	petNameLength    int
	petNamePrefix    string
	petNameSeparator string

	randSeed int64

	flags *flag.FlagSet
	help  string
}

func newKVCommand(ui cli.Ui) cli.Command {
	c := &kvCommand{
		config: kv.DefaultConfig(),
		ui:     ui,
		keyType: stringChoiceValue{
			choices: []string{keyTypePetName},
			value:   keyTypePetName,
		},
		valueType: stringChoiceValue{
			choices: []string{valueTypeRandomB64},
			value:   valueTypeRandomB64,
		},
	}

	flags := flag.NewFlagSet("", flag.ContinueOnError)

	flags.Int64Var(&c.randSeed, "seed", 0, "Value to use to seed the pseudo-random number generator with instead of the current time")
	flags.IntVar(&c.config.NumEntries, "num-entries", 1024, "Number of KV entries to generate")
	flags.IntVar(&c.randValueMinSize, "rand-value-min", 64, "Minimum byte size of random KV values to create")
	flags.IntVar(&c.randValueMaxSize, "rand-value-max", 1024, "Maximum byte size of random KV values to create")
	flags.IntVar(&c.petNameLength, "pet-name-len", 3, "Length in words of the randomly generated KV key pet names")
	flags.StringVar(&c.petNamePrefix, "pet-name-prefix", "", "Prefix for KV key pet names")
	flags.StringVar(&c.petNameSeparator, "pet-name-separator", "-", "Separator char for KV key pet names")

	c.flags = flags
	c.help = genUsage(`Usage: consul-data-gen kv [OPTIONS] [output path]
	
	Generate KV data for consul.
	
	By default the generated output is sent to the console but
	an optional output path may be used to cause it to be written
	to a file`, c.flags)

	return c
}

func (c *kvCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to parse command line arguments: %v", err))
		return 1
	}

	args = c.flags.Args()

	if c.randSeed == 0 {
		c.randSeed = time.Now().UnixNano()
	}
	rand.Seed(c.randSeed)

	switch c.keyType.value {
	case keyTypePetName:
		if c.petNameLength < 1 {
			c.ui.Error(fmt.Sprintf("Invalid pet name length: %v", c.petNameLength))
			return 1
		}
		c.config.KeyGen = kv.PetNameKeyGenerator(c.petNamePrefix, c.petNameLength, c.petNameSeparator)
	}

	switch c.valueType.value {
	case valueTypeRandomB64:
		if c.randValueMinSize < 1 {
			c.ui.Error(fmt.Sprintf("Invalid random value minimum size: %v. Value must be a positive non-zero integer", c.randValueMinSize))
			return 1
		}
		if c.randValueMaxSize < c.randValueMinSize {
			c.ui.Error(fmt.Sprintf("Invalid random value max size: %v. Value must be greater than or equal to the minimum size of %v", c.randValueMaxSize, c.randValueMinSize))
			return 1
		}

		c.config.ValueGen = kv.RandomValueGenerator(c.randValueMinSize, c.randValueMaxSize)

	}

	data, err := kv.Generate(c.config)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to generate KV data: %v\n", err))
		return 1
	}

	serialized, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to serialize KV data: %v\n", err))
		return 1
	}

	if len(args) > 0 {
		if err := ioutil.WriteFile(args[0], serialized, 0644); err != nil {
			c.ui.Error(fmt.Sprintf("Failed to write serialized KV data to %q: %v", args[0], err))
			return 1
		}
		c.ui.Info(fmt.Sprintf("KV data written to %s", args[0]))
	} else {
		c.ui.Info(string(serialized))
	}

	return 0
}

func (c *kvCommand) Synopsis() string {
	return "Generate KV data for Consul"
}

func (c *kvCommand) Help() string {
	return c.help
}
