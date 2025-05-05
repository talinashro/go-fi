// examples/orchestrator/main.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/talinashro/faultfabric/sdk"
)

func main() {
	// allow override of control-server address
	addr := flag.String("ctrl-addr", getEnv("CTRL_ADDR", "0.0.0.0:8081"),
		"control server bind address")
	flag.Parse()

	// start control server
	go startControlServer(*addr)

	// simulate your workflow
	ctx := context.Background()
	fmt.Println("→ Starting CreateDB workflow")
	if err := createDBWorkflow(ctx); err != nil {
		log.Fatalf("workflow failed: %v", err)
	}
	fmt.Println("✓ workflow completed successfully")
}

func createDBWorkflow(ctx context.Context) error {
	// EC2 step with up to 2 retries
	for attempt := 1; attempt <= 2; attempt++ {
		log.Printf("[EC2] attempt %d", attempt)
		if err := CreateEC2(ctx); err != nil {
			log.Printf("[EC2] error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}

	// Storage step, no injection by default
	if err := CreateStorage(ctx); err != nil {
		return fmt.Errorf("storage failed: %w", err)
	}
	return nil
}

func CreateEC2(ctx context.Context) error {
	if sdk.Inject("EC2") {
		return errors.New("injected EC2 failure")
	}
	log.Println("[EC2] real CreateEC2 succeeded")
	return nil
}

func CreateStorage(ctx context.Context) error {
	if sdk.Inject("STORAGE") {
		return errors.New("injected Storage failure")
	}
	log.Println("[Storage] real CreateStorage succeeded")
	return nil
}

func startControlServer(addr string) {
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		cnt, _ := strconv.Atoi(r.URL.Query().Get("count"))
		sdk.SetFailures(key, cnt)
		w.Write([]byte("OK\n"))
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(sdk.Status())
	})
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		sdk.Reset()
		w.Write([]byte("OK\n"))
	})
	log.Printf("control server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getEnv(key, dflt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return dflt
}
