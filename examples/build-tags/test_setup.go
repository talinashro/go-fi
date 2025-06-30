// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

//go:build testing

package main

import (
	"github.com/talinashro/go-fi/faultinject"
	"log"
)

func init() {
	// Only runs in test builds
	log.Println("Loading fault injection configuration...")
	if err := faultinject.LoadSpec("test-faults.yaml"); err != nil {
		log.Printf("Failed to load fault spec: %v", err)
		return
	}

	// Configure some test failures
	faultinject.SetFailures("user-create", 2)

	log.Printf("Fault injection configured: %v", faultinject.Status())
}
