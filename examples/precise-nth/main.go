package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/talinashro/go-fi"
)

func main() {
	spec := flag.String("spec", "faults.yaml", "path to faults.yaml")
	flag.Parse()

	// Load both first-N and precise-N rules
	if err := faultinject.LoadSpec(*spec); err != nil {
		log.Fatalf("LoadSpec error: %v", err)
	}
	log.Printf("Loaded fault spec: %+v\n", faultinject.Status())

	// Demonstrate create-primary-ec2 failing only on the 3rd call
	for i := 1; i <= 5; i++ {
		if err := createPrimaryNode(context.Background()); err != nil {
			fmt.Printf("Attempt %d: ERROR: %v\n", i, err)
		} else {
			fmt.Printf("Attempt %d: SUCCESS\n", i)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Demonstrate create-storage failing on first call only
	fmt.Println()
	for i := 1; i <= 3; i++ {
		if err := createStorage(context.Background()); err != nil {
			fmt.Printf("Storage Attempt %d: ERROR: %v\n", i, err)
		} else {
			fmt.Printf("Storage Attempt %d: SUCCESS\n", i)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func createPrimaryNode(ctx context.Context) error {
	key := "create-primary-ec2"
	if faultinject.Inject(key) {
		return fmt.Errorf("injected failure on %s", key)
	}
	// real logic would go here
	return nil
}

func createStorage(ctx context.Context) error {
	key := "create-storage"
	if faultinject.Inject(key) {
		return fmt.Errorf("injected failure on %s", key)
	}
	return nil
}
