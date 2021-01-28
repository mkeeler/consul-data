package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mkeeler/consul-data/generate"
)

func loadData(path string) (*generate.Data, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read data from %s: %w", path, err)
	}

	var data generate.Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON data from %s: %w", path, err)
	}

	return &data, nil
}
