package generate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mkeeler/consul-data/generate/catalog"
	"github.com/mkeeler/consul-data/generate/kv"
)

type Config struct {
	KV      kv.UserConfig
	Catalog catalog.UserConfig
}

type Data struct {
	KV      kv.KV
	Catalog catalog.Catalog
}

func GenerateAll(conf Config) (*Data, error) {
	kvConf, err := conf.KV.ToGeneratorConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to setup catalog config: %w", err)
	}

	kvData, err := kv.Generate(kvConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate KV data: %w", err)
	}

	catalogConf, err := conf.Catalog.ToGeneratorConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to setup catalog config: %w", err)
	}

	catalogData, err := catalog.Generate(catalogConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate catalog data: %w", err)
	}

	return &Data{
		KV:      kvData,
		Catalog: catalogData,
	}, nil
}

func ParseConfig(path string) (Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to read config file (%s): %w", path, err)
	}

	conf := Config{}
	if err := json.Unmarshal(data, &conf); err != nil {
		return Config{}, fmt.Errorf("Failed to parse JSON config (%s): %w", path, err)
	}

	return conf, nil
}

func DefaultConfig() Config {
	return Config{
		KV:      kv.DefaultUserConfig(),
		Catalog: catalog.DefaultUserConfig(),
	}
}
