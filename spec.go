// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package faultinject

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Spec defines the simple DSL for setting injection counts.
type Spec struct {
	Failures map[string]int `yaml:"failures"`
}

// LoadSpec loads a YAML file and applies its failure counts.
func LoadSpec(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg Spec
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	Reset()
	for k, v := range cfg.Failures {
		SetFailures(k, v)
	}
	return nil
}
