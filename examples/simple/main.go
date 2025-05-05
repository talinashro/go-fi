// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/talinashro/go-fi"
)

func main() {
	spec := flag.String("spec", "faults.yaml", "path to faults.yaml")
	flag.Parse()

	if err := faultinject.LoadSpec(*spec); err != nil {
		log.Fatalf("LoadSpec: %v", err)
	}
	log.Println("Loaded faults:", faultinject.Status())

	for i := 1; i <= 5; i++ {
		log.Printf("attempting to create an EC2 %d", i)
		if err := createEC2(context.Background()); err != nil {
			log.Println("error while attempting to create an EC2:", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	log.Println("Done.")
}

func createEC2(ctx context.Context) error {
	if faultinject.Inject("create-ec2") {
		return errors.New("injected a failure for EC2 creation")
	}
	log.Println("createEC2 succeeded")
	return nil
}
