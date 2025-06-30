package faultinject

import (
	"os"
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
			content:     `failures:\n  api-fault: 5\n  db-fault: 3`,
			expectError: false,
			expected: map[string]int{
				"api-fault": 5,
				"db-fault":  3,
			},
		},
		{
			name:        "valid precise failures content",
			content:     `precise-failures:\n  precise-api: 10\n  precise-db: 7`,
			expectError: false,
			expected: map[string]int{
				"precise-api": 10,
				"precise-db":  7,
			},
		},
		{
			name:        "both failure types",
			content:     `failures:\n  api-fault: 5\nprecise-failures:\n  precise-api: 10`,
			expectError: false,
			expected: map[string]int{
				"api-fault":  5,
				"precise-api": 10,
			},
		},
		{
			name:        "empty content",
			content:     ``,
			expectError: false,
			expected:    map[string]int{},
		},
		{
			name:        "invalid YAML",
			content:     `failures:\n  api-fault: [invalid`,
			expectError: true,
			expected:    map[string]int{},
		},
		{
			name:        "non-integer values",
			content:     `failures:\n  api-fault: "string"`,
			expectError: true,
			expected:    map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()

			err := loadSpecContent([]byte(tt.content))

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check if failures were loaded correctly
			if !tt.expectError {
				for key, expectedCount := range tt.expected {
					if failures[key] != expectedCount && preciseFailures[key] != expectedCount {
						t.Errorf("Expected %s to have count %d, but got failures[%s]=%d, preciseFailures[%s]=%d",
							key, expectedCount, key, failures[key], key, preciseFailures[key])
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
		content1 := `failures:
  api-fault: 5
  db-fault: 3`
		filename1 := "test-multi1.yaml"
		err := os.WriteFile(filename1, []byte(content1), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename1)

		// Create second file
		content2 := `precise-failures:
  precise-api: 10
  precise-db: 7`
		filename2 := "test-multi2.yaml"
		err = os.WriteFile(filename2, []byte(content2), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename2)

		// Load first file
		err = LoadSpec(filename1)
		if err != nil {
			t.Errorf("Failed to load first file: %v", err)
		}

		// Load second file
		err = LoadSpec(filename2)
		if err != nil {
			t.Errorf("Failed to load second file: %v", err)
		}

		// Verify both files were loaded
		if failures["api-fault"] != 5 {
			t.Errorf("Expected api-fault to be 5, got %d", failures["api-fault"])
		}
		if failures["db-fault"] != 3 {
			t.Errorf("Expected db-fault to be 3, got %d", failures["db-fault"])
		}
		if preciseFailures["precise-api"] != 10 {
			t.Errorf("Expected precise-api to be 10, got %d", preciseFailures["precise-api"])
		}
		if preciseFailures["precise-db"] != 7 {
			t.Errorf("Expected precise-db to be 7, got %d", preciseFailures["precise-db"])
		}
	})

	t.Run("load file with invalid content after valid file", func(t *testing.T) {
		resetState()

		// Create valid file
		content1 := `failures:
  api-fault: 5`
		filename1 := "test-valid.yaml"
		err := os.WriteFile(filename1, []byte(content1), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename1)

		// Create invalid file
		content2 := `failures:
  api-fault: "invalid"`
		filename2 := "test-invalid.yaml"
		err = os.WriteFile(filename2, []byte(content2), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename2)

		// Load valid file
		err = LoadSpec(filename1)
		if err != nil {
			t.Errorf("Failed to load valid file: %v", err)
		}

		// Load invalid file should fail
		err = LoadSpec(filename2)
		if err == nil {
			t.Error("Expected error when loading invalid file, but got none")
		}

		// Verify valid file content is still there
		if failures["api-fault"] != 5 {
			t.Errorf("Expected api-fault to still be 5, got %d", failures["api-fault"])
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

		if failures["api-fault"] != 5 {
			t.Errorf("Expected api-fault to be 5, got %d", failures["api-fault"])
		}
		if failures["db-fault"] != 3 {
			t.Errorf("Expected db-fault to be 3, got %d", failures["db-fault"])
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

		if failures["zero-fault"] != 0 {
			t.Errorf("Expected zero-fault to be 0, got %d", failures["zero-fault"])
		}
		if failures["normal-fault"] != 5 {
			t.Errorf("Expected normal-fault to be 5, got %d", failures["normal-fault"])
		}
	})

	t.Run("YAML with negative values", func(t *testing.T) {
		resetState()

		content := `failures:
  negative-fault: -1
  positive-fault: 5`
		filename := "test-negative.yaml"
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(filename)

		err = LoadSpec(filename)
		if err != nil {
			t.Errorf("Failed to load file with negative values: %v", err)
		}

		if failures["negative-fault"] != -1 {
			t.Errorf("Expected negative-fault to be -1, got %d", failures["negative-fault"])
		}
		if failures["positive-fault"] != 5 {
			t.Errorf("Expected positive-fault to be 5, got %d", failures["positive-fault"])
		}
	})
} 