package catalog

import (
	"fmt"

	"github.com/mkeeler/consul-data/generate/generators"
)

type NodeType string

const (
	NodeTypePetName NodeType = "pet-name"

	DefaultNodeType = NodeTypePetName
)

type ServiceType string

const (
	ServiceTypePetName ServiceType = "pet-name"

	DefaultServiceType = ServiceTypePetName
)

type MetaKeyType string

const (
	MetaKeyTypePetName MetaKeyType = "pet-name"

	DefaultMetaKeyType = MetaKeyTypePetName
)

type MetaValueType string

const (
	MetaValueTypeRandomB64 MetaValueType = "random-b64"

	DefaultMetaValueType = MetaValueTypeRandomB64
)

type AddressType string

const (
	AddressTypeRandomTesting AddressType = "random-testing"

	DefaultAddressType = AddressTypeRandomTesting
)

type UserConfig struct {
	NumNodes               int
	MinServicesPerNode     int
	MaxServicesPerNode     int
	MinInstancesPerService int
	MaxInstancesPerService int
	MinMetaPerNode         int
	MaxMetaPerNode         int
	MinMetaPerService      int
	MaxMetaPerService      int

	NodeType      NodeType
	ServiceType   ServiceType
	AddressType   AddressType
	MetaKeyType   MetaKeyType
	MetaValueType MetaValueType

	NodePetNames       PetNameUserConfig
	ServicePetNames    PetNameUserConfig
	MetaKeyPetNames    PetNameUserConfig
	MetaValueRandomB64 RandomB64UserConfig
}

func (c *UserConfig) ToGeneratorConfig() (Config, error) {
	c.Normalize()

	conf := Config{
		NumNodes:               c.NumNodes,
		MinServicesPerNode:     c.MinServicesPerNode,
		MaxServicesPerNode:     c.MaxServicesPerNode,
		MinInstancesPerService: c.MinInstancesPerService,
		MaxInstancesPerService: c.MaxInstancesPerService,
		MinMetaPerNode:         c.MinMetaPerNode,
		MaxMetaPerNode:         c.MaxMetaPerNode,
		MinMetaPerService:      c.MinMetaPerService,
		MaxMetaPerService:      c.MaxMetaPerService,
	}

	switch c.NodeType {
	case NodeTypePetName:
		conf.NodeGen = c.NodePetNames.Generator()
	default:
		return Config{}, fmt.Errorf("Invalid node type: %s", c.NodeType)
	}

	switch c.ServiceType {
	case ServiceTypePetName:
		conf.ServiceGen = c.ServicePetNames.Generator()
	default:
		return Config{}, fmt.Errorf("Invalid service type: %s", c.ServiceType)
	}

	switch c.MetaKeyType {
	case MetaKeyTypePetName:
		conf.MetaKeyGen = c.MetaKeyPetNames.Generator()
	default:
		return Config{}, fmt.Errorf("Invalid meta key type: %s", c.MetaKeyType)
	}

	switch c.MetaValueType {
	case MetaValueTypeRandomB64:
		conf.MetaValueGen = c.MetaValueRandomB64.Generator()
	default:
		return Config{}, fmt.Errorf("Invalid meta value type: %s", c.MetaValueType)
	}

	switch c.AddressType {
	case AddressTypeRandomTesting:
		conf.AddressGen = generators.RandomTestingIPGenerator()
	default:
		return Config{}, fmt.Errorf("Invalid address type: %s", c.MetaValueType)
	}

	return conf, nil
}

func (c *UserConfig) Normalize() {
	if c.NumNodes < 0 {
		c.NumNodes = DefaultNumNodes
	}

	if c.MinServicesPerNode <= 0 {
		c.MinServicesPerNode = DefaultMinServicesPerNode
	}

	if c.MaxServicesPerNode <= 0 {
		c.MaxServicesPerNode = DefaultMaxServicesPerNode
	}

	if c.MaxServicesPerNode < c.MinServicesPerNode {
		c.MaxServicesPerNode = c.MinServicesPerNode
	}

	if c.MinMetaPerNode <= 0 {
		c.MinMetaPerNode = DefaultMinMetaPerNode
	}

	if c.MaxMetaPerNode <= 0 {
		c.MaxMetaPerNode = DefaultMaxMetaPerNode
	}

	if c.MaxMetaPerNode < c.MinMetaPerNode {
		c.MaxMetaPerNode = c.MinMetaPerNode
	}

	if c.MinMetaPerService <= 0 {
		c.MinMetaPerService = DefaultMinMetaPerService
	}

	if c.MaxMetaPerService <= 0 {
		c.MaxMetaPerService = DefaultMaxMetaPerService
	}

	if c.MaxMetaPerService < c.MinMetaPerService {
		c.MaxMetaPerService = c.MinMetaPerService
	}

	if c.NodeType == "" {
		c.NodeType = NodeTypePetName
	}

	if c.ServiceType == "" {
		c.ServiceType = ServiceTypePetName
	}

	if c.MetaKeyType == "" {
		c.MetaKeyType = MetaKeyTypePetName
	}

	if c.MetaValueType == "" {
		c.MetaValueType = MetaValueTypeRandomB64
	}

	if c.AddressType == "" {
		c.AddressType = AddressTypeRandomTesting
	}

	c.NodePetNames.Normalize()
	c.ServicePetNames.Normalize()
	c.MetaKeyPetNames.Normalize()
	c.MetaValueRandomB64.Normalize()
}

func DefaultUserConfig() UserConfig {
	return UserConfig{
		NumNodes:           DefaultNumNodes,
		MinServicesPerNode: DefaultMinServicesPerNode,
		MaxServicesPerNode: DefaultMaxServicesPerNode,
		MinMetaPerNode:     DefaultMinMetaPerNode,
		MaxMetaPerNode:     DefaultMaxMetaPerNode,
		MinMetaPerService:  DefaultMinMetaPerService,
		MaxMetaPerService:  DefaultMaxMetaPerService,
		NodeType:           DefaultNodeType,
		ServiceType:        DefaultServiceType,
		AddressType:        DefaultAddressType,
		MetaKeyType:        DefaultMetaKeyType,
		MetaValueType:      DefaultMetaValueType,
		NodePetNames:       DefaultPetNameUserConfig(),
		ServicePetNames:    DefaultPetNameUserConfig(),
		MetaKeyPetNames:    DefaultPetNameUserConfig(),
		MetaValueRandomB64: DefaultRandomB64UserConfig(),
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
