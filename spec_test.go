package faultinject

import (
	"os"
	"strings"
	"testing"
)

func TestLoadSpec(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name        string
		filename    string
		content     string
		expectError bool
		setup       func() string
		cleanup     func(string)
	}{
		{
			name:        "valid YAML with failures",
			expectError: false,
			setup: func() string {
				content := `failures:
  api-fault: 5
  db-fault: 3
  timeout-fault: 1`
				filename := "test-valid.yaml"
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filename
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:        "valid YAML with precise failures",
			expectError: false,
			setup: func() string {
				content := `precise-failures:
  precise-api: 10
  precise-db: 7`
				filename := "test-precise.yaml"
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filename
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:        "valid YAML with both failure types",
			expectError: false,
			setup: func() string {
				content := `failures:
  api-fault: 5
  db-fault: 3
precise-failures:
  precise-api: 10
  precise-db: 7`
				filename := "test-both.yaml"
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filename
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:        "empty YAML file",
			expectError: false,
			setup: func() string {
				content := ``
				filename := "test-empty.yaml"
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filename
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:        "invalid YAML syntax",
			expectError: true,
			setup: func() string {
				content := `failures:
  api-fault: 5
  db-fault: [invalid: yaml`
				filename := "test-invalid.yaml"
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filename
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:        "non-existent file",
			expectError: true,
			setup: func() string {
				return "non-existent-file.yaml"
			},
		},
		{
			name:        "YAML with non-integer values",
			expectError: true,
			setup: func() string {
				content := `failures:
  api-fault: "not a number"
  db-fault: 3.14`
				filename := "test-non-int.yaml"
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filename
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()

			var filename string
			if tt.setup != nil {
				filename = tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup(filename)
			}

			err := LoadSpec(filename)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestLoadSpecContent(t *testing.T) {
	// Reset state before each test
	resetState()

	tests := []struct {
		name        string
		content     string
		expectError bool
		expected    map[string]int
	}{
		{
			name:        "valid failures content",
			content:     "failures:\n  api-fault: 5\n  db-fault: 3",
			expectError: false,
			expected: map[string]int{
				"api-fault": 5,
				"db-fault":  3,
			},
		},
		{
			name:        "valid precise failures content",
			content:     "precise-failures:\n  precise-api: 10\n  precise-db: 7",
			expectError: false,
			expected:    map[string]int{}, // Status() does not expose precise failures
		},
		{
			name:        "both failure types",
			content:     "failures:\n  api-fault: 5\nprecise-failures:\n  precise-api: 10",
			expectError: false,
			expected: map[string]int{
				"api-fault": 5,
			},
		},
		{
			name:        "empty content",
			content:     "",
			expectError: false,
			expected:    map[string]int{},
		},
		{
			name:        "invalid YAML",
			content:     "failures:\n  api-fault: [invalid",
			expectError: true,
			expected:    map[string]int{},
		},
		{
			name:        "non-integer values",
			content:     "failures:\n  api-fault: \"string\"",
			expectError: true,
			expected:    map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()

			// Fix: Replace \n with real newlines for valid YAML
			content := strings.ReplaceAll(tt.content, "\\n", "\n")

			filename := "temp-spec.yaml"
			err := os.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(filename)

			err = LoadSpec(filename)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check if failures were loaded correctly (only first-N failures)
			if !tt.expectError {
				status := Status()
				for key, expectedCount := range tt.expected {
					if status[key] != expectedCount {
						t.Errorf("Expected %s to have count %d, but got %d",
							key, expectedCount, status[key])
					}
				}
			}
		})
	}
}

func TestLoadSpecMultipleFiles(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("load multiple valid files", func(t *testing.T) {
		resetState()

		// Create first file
		content1 := "failures:\n  api-fault: 5\n  db-fault: 3"
		filename1 := "test-multi1.yaml"
		err := os.WriteFile(filename1, []byte(strings.ReplaceAll(content1, "\\n", "\n")), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename1)

		// Create second file
		content2 := "failures:\n  only-fault: 42"
		filename2 := "test-multi2.yaml"
		err = os.WriteFile(filename2, []byte(strings.ReplaceAll(content2, "\\n", "\n")), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename2)

		// Load first file
		err = LoadSpec(filename1)
		if err != nil {
			t.Errorf("Failed to load first file: %v", err)
		}

		// Check first file's failures
		status := Status()
		if status["api-fault"] != 5 {
			t.Errorf("Expected api-fault to be 5, got %d", status["api-fault"])
		}
		if status["db-fault"] != 3 {
			t.Errorf("Expected db-fault to be 3, got %d", status["db-fault"])
		}

		// Load second file (should reset state)
		err = LoadSpec(filename2)
		if err != nil {
			t.Errorf("Failed to load second file: %v", err)
		}

		// After loading the second file, only its failures should be present
		status = Status()
		if len(status) != 1 || status["only-fault"] != 42 {
			t.Errorf("Expected only-fault to be 42 and only one fault present, got %+v", status)
		}
	})

	t.Run("load file with invalid content after valid file", func(t *testing.T) {
		resetState()

		// Create valid file
		content1 := "failures:\n  api-fault: 5"
		filename1 := "test-valid.yaml"
		err := os.WriteFile(filename1, []byte(strings.ReplaceAll(content1, "\\n", "\n")), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename1)

		// Create invalid file
		content2 := "failures:\n  api-fault: \"invalid\""
		filename2 := "test-invalid.yaml"
		err = os.WriteFile(filename2, []byte(strings.ReplaceAll(content2, "\\n", "\n")), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename2)

		// Load valid file
		err = LoadSpec(filename1)
		if err != nil {
			t.Errorf("Failed to load valid file: %v", err)
		}

		// Load invalid file should fail, but state should remain as after last successful load
		err = LoadSpec(filename2)
		if err == nil {
			t.Error("Expected error when loading invalid file, but got none")
		}

		// Verify valid file content is still there
		status := Status()
		if len(status) != 1 || status["api-fault"] != 5 {
			t.Errorf("Expected api-fault to still be 5 and only one fault present, got %+v", status)
		}
	})
}

func TestLoadSpecEdgeCases(t *testing.T) {
	// Reset state before each test
	resetState()

	t.Run("YAML with comments", func(t *testing.T) {
		resetState()

		content := `# This is a comment
failures:
  api-fault: 5  # Another comment
  db-fault: 3`
		filename := "test-comments.yaml"
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename)

		err = LoadSpec(filename)
		if err != nil {
			t.Errorf("Failed to load file with comments: %v", err)
		}

		status := Status()
		if status["api-fault"] != 5 {
			t.Errorf("Expected api-fault to be 5, got %d", status["api-fault"])
		}
		if status["db-fault"] != 3 {
			t.Errorf("Expected db-fault to be 3, got %d", status["db-fault"])
		}
	})

	t.Run("YAML with zero values", func(t *testing.T) {
		resetState()

		content := `failures:
  zero-fault: 0
  normal-fault: 5`
		filename := "test-zero.yaml"
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename)

		err = LoadSpec(filename)
		if err != nil {
			t.Errorf("Failed to load file with zero values: %v", err)
		}

		status := Status()
		if status["zero-fault"] != 0 {
			t.Errorf("Expected zero-fault to be 0, got %d", status["zero-fault"])
		}
		if status["normal-fault"] != 5 {
			t.Errorf("Expected normal-fault to be 5, got %d", status["normal-fault"])
		}
	})

	t.Run("YAML with negative values", func(t *testing.T) {
		resetState()

		content := "failures:\n  negative-fault: -1\n  positive-fault: 5"
		filename := "test-negative.yaml"
		err := os.WriteFile(filename, []byte(strings.ReplaceAll(content, "\\n", "\n")), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename)

		err = LoadSpec(filename)
		if err != nil {
			t.Errorf("Failed to load file with negative values: %v", err)
		}

		status := Status()
		if len(status) != 2 || status["negative-fault"] != 0 || status["positive-fault"] != 5 {
			t.Errorf("Expected negative-fault: 0 and positive-fault: 5, got %+v", status)
		}
	})
}
