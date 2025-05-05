// Copyright 2025 Talina Shrotriya
// SPDX-License-Identifier: Apache-2.0

package faultinject

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// StartControlServer starts an HTTP server on addr with /set, /reset, /status, and optional /run.
func StartControlServer(addr string, runHandler http.HandlerFunc) {
	mux := http.NewServeMux()

	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("key")
		c, _ := strconv.Atoi(r.URL.Query().Get("count"))
		SetFailures(k, c)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		Reset()
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Status())
	})

	if runHandler != nil {
		mux.HandleFunc("/run", runHandler)
	}

	go http.ListenAndServe(addr, mux)
}
