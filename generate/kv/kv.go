package kv

import (
	"fmt"

	"github.com/mkeeler/consul-data/generate/generators"
)

var (
	DefaultKeyGenerator   = generators.PetNameGenerator("", 3, "-")
	DefaultValueGenerator = generators.RandomB64Generator(64, 1024)
)

// Value is the value type of the KV mapping
type Value struct {
	Datacenter string `json:",omitempty"`
	Token      string `json:",omitempty"`
	Value      string
	Namespace  string `json:",omitempty"`
	Flags      uint   `json:",omitempty"`
}

// KV is the output format of the generated KV data before serializing to JSON.
// This will result in a JSON object that looks like the following:
//   {
//      "<key 1>": {
//         "Value": "<data>",
//         "Namespace": "<optional>",
//         "Datacenter": "<optional>",
//         "Token": "<optional>",
//         "Flags": <optional - uint>,
//      },
//      ...
//   }
//
// All fields of each object may be omitted if they are empty except for the `Value` field
type KV map[string]Value

// Config is all the configuration necessary for creating KV data
type Config struct {
	NumEntries int
	KeyGen     generators.StringGenerator
	ValueGen   generators.StringGenerator
}

// DefaultConfig returns a config with all the defaults filled in.
func DefaultConfig() Config {
	return Config{
		NumEntries: 1024,
		KeyGen:     DefaultKeyGenerator,
		ValueGen:   DefaultValueGenerator,
	}
}

func genNewKey(existing KV, gen generators.StringGenerator) (string, error) {
	for {
		key, err := gen()
		if err != nil {
			return "", fmt.Errorf("Failed to generate KV Key: %w", err)
		}

		if _, found := existing[key]; !found {
			return key, nil
		}
	}
}

// Generate will generate the desired number of KV entries giving the supplied config
func Generate(conf Config) (KV, error) {
	if conf.KeyGen == nil {
		conf.KeyGen = DefaultKeyGenerator
	}

	if conf.ValueGen == nil {
		conf.ValueGen = DefaultValueGenerator
	}

	if conf.NumEntries < 1 {
		return make(KV), nil
	}

	data := make(KV)

	for i := 0; i < conf.NumEntries; i++ {

		key, err := genNewKey(data, conf.KeyGen)
		if err != nil {
			return nil, err
		}

		value, err := conf.ValueGen()
		if err != nil {
			return nil, fmt.Errorf("Failed to generate KV Value: %w", err)
		}

		data[key] = Value{Value: value}
	}

	return data, nil
}
