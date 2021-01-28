package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/cli"
	"github.com/mkeeler/consul-data/generate"
)

type pushCommand struct {
	ui         cli.Ui
	configPath string
	dataPath   string
	outputPath string
	randSeed   int64
	parallel   int
	quiet      bool

	flags *flag.FlagSet
	http  *HTTPFlags
	help  string
}

func newPushCommand(ui cli.Ui) cli.Command {
	c := &pushCommand{
		ui: ui,
	}

	flags := flag.NewFlagSet("", flag.ContinueOnError)

	flags.BoolVar(&c.quiet, "quiet", false, "Whether to suppress output of handling of individual resources")
	flags.IntVar(&c.parallel, "parallel", 1, "Number of concurrent requests that can be made")
	flags.Int64Var(&c.randSeed, "seed", 0, "Value to use to seed the pseudo-random number generator with instead of the current time")
	flags.StringVar(&c.configPath, "config", "", "Path to the configuration to use for generating data")
	flags.StringVar(&c.dataPath, "data", "", "Path to data generated by consul-data generate to use as the data source instead of generating new data")
	flags.StringVar(&c.outputPath, "output", "", "Path to output the data file to if we generated it instead of loading it in")

	c.http = &HTTPFlags{}
	c.http.MergeAll(flags)

	c.flags = flags
	c.help = genUsage(`Usage: consul-data push [OPTIONS]
	
	Push data to Consul
	
	By default this command will generate the random data and push it
	to Consul. The data can be pregenerated if the -data flag is used
	to specify a file which should be in the format outputted by the
	consul-data generate command`, c.flags)

	return c
}

func (c *pushCommand) generateData() (*generate.Data, error) {
	if c.randSeed == 0 {
		c.randSeed = time.Now().UnixNano()
	}
	rand.Seed(c.randSeed)

	conf := generate.DefaultConfig()

	if c.configPath != "" {
		var err error
		conf, err = generate.ParseConfig(c.configPath)
		if err != nil {
			return nil, err
		}
	}

	return generate.GenerateAll(conf)
}

func (c *pushCommand) loadData() (*generate.Data, error) {
	raw, err := ioutil.ReadFile(c.dataPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read data from %s: %w", c.dataPath, err)
	}

	var data generate.Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON data from %s: %w", c.dataPath, err)
	}

	return &data, nil
}

func (c *pushCommand) getData() (*generate.Data, error) {
	if c.dataPath != "" {
		return c.loadData()
	}
	return c.generateData()
}

func (c *pushCommand) pushData(data *generate.Data) error {
	client, err := c.http.APIClient()
	if err != nil {
		return fmt.Errorf("Failed to create Consul API client: %w", err)
	}

	resources := 0

	c.ui.Info("Pushing KV data to Consul")
	kv := client.KV()
	for key, value := range data.KV {
		if !c.quiet {
			c.ui.Output(fmt.Sprintf("   Key: %s", key))
		}
		pair := api.KVPair{
			Key:       key,
			Value:     []byte(value.Value),
			Flags:     uint64(value.Flags),
			Namespace: value.Namespace,
		}

		opts := api.WriteOptions{
			Datacenter: value.Datacenter,
			Token:      value.Token,
		}

		_, err := kv.Put(&pair, &opts)
		if err != nil {
			return fmt.Errorf("Failed to push key %s: %w", key, err)
		}
		resources += 1
	}
	c.ui.Info("Finished pushing KV data to Consul")

	c.ui.Info("Pushing Catalog data to Consul")
	catalog := client.Catalog()
	for _, node := range data.Catalog {
		if !c.quiet {
			c.ui.Output(fmt.Sprintf("   Node: %s", node.Name))
		}

		nodeRegistration := api.CatalogRegistration{
			ID:         node.ID,
			Node:       node.Name,
			Address:    node.Address,
			NodeMeta:   node.Meta,
			Datacenter: node.Datacenter,
		}

		_, err := catalog.Register(&nodeRegistration, nil)
		if err != nil {
			return fmt.Errorf("Failed to push Node %s: %w", node.Name, err)
		}

		resources += 1

		for _, service := range node.Services {
			if !c.quiet {
				c.ui.Output(fmt.Sprintf("      Service: %s", service.Name))
			}
			serviceRegistration := api.CatalogRegistration{
				ID:             node.ID,
				Node:           node.Name,
				Datacenter:     node.Datacenter,
				SkipNodeUpdate: true,
				Service: &api.AgentService{
					ID:      service.ID,
					Service: service.Name,
					Address: service.Address,
					Port:    service.Port,
					Meta:    service.Meta,
				},
			}

			_, err := catalog.Register(&serviceRegistration, nil)
			if err != nil {
				return fmt.Errorf("Failed to push Service %s for node %s: %w", service.Name, node.Name, err)
			}

			resources += 1
		}
	}
	c.ui.Info("Finished pushing Catalog data to Consul")

	c.ui.Info(fmt.Sprintf("Total Resources Created: %d", resources))
	return nil
}

func (c *pushCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to parse command line arguments: %v", err))
		return 1
	}

	if c.dataPath != "" && c.outputPath != "" {
		c.ui.Error("Cannot specify both a data path and an output path")
		return 1
	}

	data, err := c.getData()
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	if err := c.pushData(data); err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	if c.outputPath != "" {
		serialized, err := json.MarshalIndent(data, "", "   ")
		if err != nil {
			c.ui.Error(fmt.Sprintf("Failed to serialize Consul data: %v", err))
			return 1
		}

		if err := ioutil.WriteFile(c.outputPath, serialized, 0644); err != nil {
			c.ui.Error(fmt.Sprintf("Failed to write serialized Consul data to %q: %v", c.outputPath, err))
			return 1
		}
		c.ui.Info(fmt.Sprintf("Consul data written to %s", args[0]))
	}

	return 0
}

func (c *pushCommand) Synopsis() string {
	return "Push data to Consul"
}

func (c *pushCommand) Help() string {
	return c.help
}
