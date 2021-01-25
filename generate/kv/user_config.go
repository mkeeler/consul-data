package kv

import (
	"fmt"

	"github.com/mkeeler/consul-data/generate/generators"
)

type KeyType string

const (
	KeyTypePetName KeyType = "pet-name"

	DefaultKeyType = KeyTypePetName
)

type ValueType string

const (
	ValueTypeRandomB64 ValueType = "random-b64"

	DefaultValueType = ValueTypeRandomB64
)

const (
	DefaultNumEntries = 1024
)

type UserConfig struct {
	NumEntries int
	KeyType    KeyType
	ValueType  ValueType

	PetName   PetNameUserConfig
	RandomB64 RandomB64UserConfig
}

func (c *UserConfig) ToGeneratorConfig() (Config, error) {
	c.Normalize()

	conf := Config{
		NumEntries: c.NumEntries,
	}

	switch c.KeyType {
	case KeyTypePetName:
		conf.KeyGen = c.PetName.Generator()
	default:
		return Config{}, fmt.Errorf("Invalid KV generator key type: %s", c.KeyType)
	}

	switch c.ValueType {
	case ValueTypeRandomB64:
		conf.ValueGen = c.RandomB64.Generator()
	default:
		return Config{}, fmt.Errorf("Invalid KV generator key type: %s", c.KeyType)
	}

	return conf, nil
}

func (c *UserConfig) Normalize() {
	if c.NumEntries < 0 {
		c.NumEntries = DefaultNumEntries
	}

	if c.KeyType == "" {
		c.KeyType = DefaultKeyType
	}

	if c.ValueType == "" {
		c.ValueType = DefaultValueType
	}

	c.PetName.Normalize()
	c.RandomB64.Normalize()
}

func DefaultUserConfig() UserConfig {
	return UserConfig{
		NumEntries: DefaultNumEntries,
		KeyType:    DefaultKeyType,
		ValueType:  DefaultValueType,

		PetName:   DefaultPetNameUserConfig(),
		RandomB64: DefaultRandomB64UserConfig(),
	}
}

const (
	PetNameDefaultPrefix    = ""
	PetNameDefaultSegments  = 3
	PetNameDefaultSeparator = "-"
)

type PetNameUserConfig struct {
	Prefix    string
	Segments  int
	Separator string
}

func (c *PetNameUserConfig) Normalize() {
	if c.Segments < 1 {
		c.Segments = PetNameDefaultSegments
	}

	if c.Separator == "" {
		c.Separator = PetNameDefaultSeparator
	}
}

func (c *PetNameUserConfig) Generator() generators.StringGenerator {
	return generators.PetNameGenerator(c.Prefix, c.Segments, c.Separator)
}

func DefaultPetNameUserConfig() PetNameUserConfig {
	return PetNameUserConfig{
		Prefix:    PetNameDefaultPrefix,
		Segments:  PetNameDefaultSegments,
		Separator: PetNameDefaultSeparator,
	}
}

const (
	RandomB64DefaultMinSize = 64
	RandomB64DefaultMaxSize = 1024
)

type RandomB64UserConfig struct {
	MinSize int
	MaxSize int
}

func (c *RandomB64UserConfig) Normalize() {
	if c.MinSize <= 0 && c.MaxSize <= 0 {
		c.MinSize = RandomB64DefaultMinSize
		c.MaxSize = RandomB64DefaultMaxSize
	} else if c.MinSize <= 0 {
		c.MinSize = RandomB64DefaultMinSize
	} else if c.MaxSize <= 0 {
		c.MaxSize = RandomB64DefaultMaxSize
	}

	if c.MaxSize < c.MinSize {
		c.MaxSize = c.MinSize
	}
}

func (c *RandomB64UserConfig) Generator() generators.StringGenerator {
	return generators.RandomB64Generator(c.MinSize, c.MaxSize)
}

func DefaultRandomB64UserConfig() RandomB64UserConfig {
	return RandomB64UserConfig{
		MinSize: RandomB64DefaultMinSize,
		MaxSize: RandomB64DefaultMaxSize,
	}
}
