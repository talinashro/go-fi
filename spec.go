// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package faultinject

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Spec struct {
	Failures        map[string]int `yaml:"failures"`         // first-N
	PreciseFailures map[string]int `yaml:"precise-failures"` // Nth
}

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
	for k, v := range cfg.PreciseFailures {
		SetNthFailure(k, v)
	}
	return nil
}
