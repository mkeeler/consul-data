package catalog

import (
	"fmt"
	"math/rand"

	"github.com/mkeeler/consul-data/generate/generators"
)

var (
	DefaultNodeNameGenerator    = generators.PetNameGenerator("", 3, "-")
	DefaultServiceNameGenerator = generators.PetNameGenerator("", 2, "-")
	DefaultMetaKeyGenerator     = generators.PetNameGenerator("", 1, "")
	DefaultMetaValueGenerator   = generators.RandomB64Generator(64, 128)
	DefaultAddressGenerator     = generators.RandomTestingIPGenerator()

	DefaultNumNodes               = 1024
	DefaultMinServicesPerNode     = 8
	DefaultMaxServicesPerNode     = 32
	DefaultMinInstancesPerService = 1
	DefaultMaxInstancesPerService = 1
	DefaultMinMetaPerNode         = 4
	DefaultMaxMetaPerNode         = 8
	DefaultMinMetaPerService      = 4
	DefaultMaxMetaPerService      = 8
)

// Node is the representation of a node
type Node struct {
	Datacenter string `json:",omitempty"`
	Address    string
	ID         string
	Name       string
	Meta       map[string]string `json:",omitempty"`
	Services   []*Service
}

type Service struct {
	Name      string
	Instances []*ServiceInstance
}

type ServiceInstance struct {
	Name    string
	Address string
	ID      string `json:",omitempty"`
	Port    int
	Meta    map[string]string `json:",omitempty"`
}

// Catalog is the output format of the generated catalog data before serialized to JSON.
type Catalog []*Node

// Config is all the configuration necessary for creating KV data
type Config struct {
	NumNodes               int
	MinServicesPerNode     int
	MaxServicesPerNode     int
	MinInstancesPerService int
	MaxInstancesPerService int
	MinMetaPerNode         int
	MaxMetaPerNode         int
	MinMetaPerService      int
	MaxMetaPerService      int
	NodeGen                generators.StringGenerator
	ServiceGen             generators.StringGenerator
	ServiceIDGen           generators.StringGenerator
	MetaKeyGen             generators.StringGenerator
	MetaValueGen           generators.StringGenerator
	AddressGen             generators.IPGenerator
}

// DefaultConfig returns a config with all the defaults filled in.
func DefaultConfig() Config {
	return Config{
		NumNodes:               DefaultNumNodes,
		MinServicesPerNode:     DefaultMinServicesPerNode,
		MaxServicesPerNode:     DefaultMaxServicesPerNode,
		MinInstancesPerService: DefaultMinInstancesPerService,
		MaxInstancesPerService: DefaultMaxInstancesPerService,
		MinMetaPerNode:         DefaultMinMetaPerNode,
		MaxMetaPerNode:         DefaultMaxMetaPerNode,
		MinMetaPerService:      DefaultMinMetaPerService,
		MaxMetaPerService:      DefaultMaxMetaPerService,
		NodeGen:                DefaultNodeNameGenerator,
		ServiceGen:             DefaultServiceNameGenerator,
		MetaKeyGen:             DefaultMetaKeyGenerator,
		AddressGen:             DefaultAddressGenerator,
	}
}

func uniqueString(gen generators.StringGenerator, isUnique func(string) bool) (string, error) {
	for {
		value, err := gen()
		if err != nil {
			return "", err
		}

		if isUnique(value) {
			return value, nil
		}
	}
}

// when determining service ids
type generatorState struct {
	// map of map to int where the outer map's keys are node names
	// and the inner maps keys are service names. The inner map value
	// represents how many of this named service were assigned to this node
	nodeAndServiceNames map[string]map[string]int

	nodeIds map[string]struct{}
}

func (g *generatorState) initNode(name string) {
	g.nodeAndServiceNames[name] = make(map[string]int)
}

func (g *generatorState) nodeNameIsUnique(name string) bool {
	_, found := g.nodeAndServiceNames[name]
	return !found
}

func (g *generatorState) svcID(node string, svc string) string {
	services := g.nodeAndServiceNames[node]

	if services == nil {
		return svc
	}

	services[svc] += 1

	return fmt.Sprintf("%s-%d", svc, services[svc])
}

func (g *generatorState) genMeta(minEntries int, maxEntries int, keyGen generators.StringGenerator, valueGen generators.StringGenerator) (map[string]string, error) {
	numEntries := minEntries
	if minEntries < maxEntries {
		numEntries = rand.Intn(maxEntries-minEntries) + minEntries
	}

	meta := make(map[string]string)
	for i := 0; i < numEntries; i++ {
		value, err := valueGen()
		if err != nil {
			return nil, fmt.Errorf("Failed to generate meta value: %w", err)
		}

		key, err := uniqueString(keyGen, func(val string) bool {
			_, found := meta[val]
			return !found
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to generate meta key: %w", err)
		}

		meta[key] = meta[value]
	}
	return meta, nil
}

func (g *generatorState) genNodeMeta(conf Config) (map[string]string, error) {
	return g.genMeta(conf.MinMetaPerNode, conf.MaxMetaPerNode, conf.MetaKeyGen, conf.MetaValueGen)
}

func (g *generatorState) genServiceMeta(conf Config) (map[string]string, error) {
	return g.genMeta(conf.MinMetaPerService, conf.MaxMetaPerService, conf.MetaKeyGen, conf.MetaValueGen)
}

func (g *generatorState) genService(nodeName string, conf Config) (*Service, error) {
	numInstances := conf.MinInstancesPerService
	if conf.MinInstancesPerService < conf.MaxInstancesPerService {
		numInstances = rand.Intn(conf.MaxInstancesPerService-conf.MinInstancesPerService) + conf.MinInstancesPerService
	}

	svcName, err := conf.ServiceGen()
	if err != nil {
		return nil, fmt.Errorf("Failed to generate service name: %w", err)
	}

	data := make([]*ServiceInstance, 0, numInstances)

	for i := 0; i < numInstances; i++ {
		service, err := g.genServiceInstance(nodeName, svcName, conf)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate Service: %w", err)
		}

		data = append(data, service)
	}

	return &Service{
		Name:      svcName,
		Instances: data,
	}, nil
}

func (g *generatorState) genServiceInstance(nodeName string, svcName string, conf Config) (*ServiceInstance, error) {
	addr, err := conf.AddressGen()
	if err != nil {
		return nil, fmt.Errorf("Failed to generate service address: %w", err)
	}

	meta, err := g.genServiceMeta(conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate service meta: %w", err)
	}

	return &ServiceInstance{
		Name:    svcName,
		Address: addr.String(),
		ID:      g.svcID(nodeName, svcName),
		Port:    rand.Intn(65535),
		Meta:    meta,
	}, nil
}

func (g *generatorState) genServices(nodeName string, conf Config) ([]*Service, error) {
	numServices := conf.MinServicesPerNode
	if conf.MinServicesPerNode < conf.MaxServicesPerNode {
		numServices = rand.Intn(conf.MaxServicesPerNode-conf.MinServicesPerNode) + conf.MinServicesPerNode
	}

	data := make([]*Service, 0, numServices)

	for i := 0; i < numServices; i++ {
		service, err := g.genService(nodeName, conf)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate Service: %w", err)
		}

		data = append(data, service)
	}
	return data, nil
}

func (g *generatorState) genNode(conf Config) (*Node, error) {
	nodeName, err := uniqueString(conf.NodeGen, g.nodeNameIsUnique)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate node name: %w", err)
	}

	nodeID, err := uniqueString(generators.UUIDGen, func(val string) bool {
		_, found := g.nodeIds[val]
		return !found
	})

	g.initNode(nodeName)

	addr, err := conf.AddressGen()
	if err != nil {
		return nil, fmt.Errorf("Failed to generate node address: %w", err)
	}

	meta, err := g.genNodeMeta(conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate node meta: %w", err)
	}

	services, err := g.genServices(nodeName, conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate services for node: %w", err)
	}

	return &Node{
		Address:  addr.String(),
		ID:       nodeID,
		Name:     nodeName,
		Meta:     meta,
		Services: services,
	}, nil
}

// Generate will generate the desired number of KV entries giving the supplied config
func Generate(conf Config) (Catalog, error) {
	conf.normalize()

	data := make(Catalog, 0, conf.NumNodes)

	g := generatorState{nodeAndServiceNames: make(map[string]map[string]int)}

	for i := 0; i < conf.NumNodes; i++ {
		node, err := g.genNode(conf)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate Node: %w", err)
		}

		data = append(data, node)
	}

	return data, nil
}

func (c *Config) normalize() {
	if c.NodeGen == nil {
		c.NodeGen = DefaultNodeNameGenerator
	}

	if c.ServiceGen == nil {
		c.ServiceGen = DefaultServiceNameGenerator
	}

	if c.MetaKeyGen == nil {
		c.MetaKeyGen = DefaultMetaKeyGenerator
	}

	if c.MetaValueGen == nil {
		c.MetaValueGen = DefaultMetaValueGenerator
	}

	if c.NumNodes < 1 {
		c.NumNodes = 0
	}

	if c.MinServicesPerNode < 0 {
		c.MinServicesPerNode = 0
	}

	if c.MaxServicesPerNode < c.MinServicesPerNode {
		c.MaxServicesPerNode = c.MinServicesPerNode
	}

	if c.MinMetaPerNode < 0 {
		c.MinMetaPerNode = 0
	}

	if c.MaxMetaPerNode < c.MinMetaPerNode {
		c.MaxMetaPerNode = c.MinMetaPerNode
	}

	if c.MinMetaPerService < 0 {
		c.MinMetaPerService = 0
	}

	if c.MaxMetaPerService < c.MinMetaPerService {
		c.MaxMetaPerService = c.MinMetaPerService
	}
}
