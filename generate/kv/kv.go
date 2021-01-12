package kv

import (
	"encoding/base64"
	"fmt"
	"math/rand"

	petname "github.com/dustinkirkland/golang-petname"
)

var (
	DefaultKeyGenerator   = PetNameKeyGenerator("", 3, "-")
	DefaultValueGenerator = RandomValueGenerator(64, 1024)
)

// Key is the type of value used to index the KV mapping
type Key string

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
type KV map[Key]Value

// KeyGenerator is a function that can generate keys from some source
type KeyGenerator func() (Key, error)

// Value Generator is a function that can generate KV values from some source.
type ValueGenerator func() (Value, error)

// Config is all the configuration necessary for creating KV data
type Config struct {
	NumEntries int
	KeyGen     KeyGenerator
	ValueGen   ValueGenerator
}

// DefaultConfig returns a config with all the defaults filled in.
func DefaultConfig() Config {
	return Config{
		NumEntries: 1024,
		KeyGen:     DefaultKeyGenerator,
		ValueGen:   DefaultValueGenerator,
	}
}

func genNewKey(existing KV, gen KeyGenerator) (Key, error) {
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
		return nil, nil
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

		data[key] = value
	}

	return data, nil
}

// PetNameKeyGenerator will generate KV
func PetNameKeyGenerator(prefix string, words int, separator string) KeyGenerator {
	return func() (Key, error) {
		return Key(fmt.Sprintf("%s%s", prefix, petname.Generate(words, separator))), nil
	}
}

func RandomValueGenerator(minSize int, maxSize int) ValueGenerator {
	return func() (Value, error) {
		size := minSize + rand.Intn(maxSize-minSize)
		raw := make([]byte, size)
		_, err := rand.Read(raw)

		// Technically math/rand.Read is guaranteed to always return a nil error but
		// we are checking anyways just in case we switch over to something else like
		// crypto/rand where the same guarantees might not be in place.
		if err != nil {
			return Value{}, fmt.Errorf("Failed to generate random KV value: %w", err)
		}

		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
		base64.StdEncoding.Encode(encoded, raw)

		return Value{
			Value: string(encoded),
		}, nil
	}
}
