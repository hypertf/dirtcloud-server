package chaos

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/nicolas/dirtcloud/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseLatencyRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *LatencyRange
	}{
		{
			name:     "valid range",
			input:    "10-100",
			expected: &LatencyRange{Min: 10, Max: 100},
		},
		{
			name:     "same min max",
			input:    "50-50",
			expected: &LatencyRange{Min: 50, Max: 50},
		},
		{
			name:     "zero range",
			input:    "0-0",
			expected: &LatencyRange{Min: 0, Max: 0},
		},
		{
			name:     "invalid format - no dash",
			input:    "100",
			expected: nil,
		},
		{
			name:     "invalid format - multiple dashes",
			input:    "10-50-100",
			expected: nil,
		},
		{
			name:     "invalid format - non-numeric",
			input:    "abc-def",
			expected: nil,
		},
		{
			name:     "invalid range - negative min",
			input:    "-10-100",
			expected: nil,
		},
		{
			name:     "invalid range - max less than min",
			input:    "100-50",
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace handling",
			input:    " 10 - 100 ",
			expected: &LatencyRange{Min: 10, Max: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLatencyRange(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseIntList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "valid list",
			input:    "503,500,429",
			expected: []int{503, 500, 429},
		},
		{
			name:     "single item",
			input:    "500",
			expected: []int{500},
		},
		{
			name:     "with spaces",
			input:    " 503 , 500 , 429 ",
			expected: []int{503, 500, 429},
		},
		{
			name:     "mixed valid and invalid",
			input:    "503,abc,500,def,429",
			expected: []int{503, 500, 429},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only invalid values",
			input:    "abc,def,ghi",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{name: "true", envValue: "true", defaultValue: false, expected: true},
		{name: "1", envValue: "1", defaultValue: false, expected: true},
		{name: "yes", envValue: "yes", defaultValue: false, expected: true},
		{name: "on", envValue: "on", defaultValue: false, expected: true},
		{name: "TRUE", envValue: "TRUE", defaultValue: false, expected: true},
		{name: "false", envValue: "false", defaultValue: true, expected: false},
		{name: "0", envValue: "0", defaultValue: true, expected: false},
		{name: "no", envValue: "no", defaultValue: true, expected: false},
		{name: "off", envValue: "off", defaultValue: true, expected: false},
		{name: "FALSE", envValue: "FALSE", defaultValue: true, expected: false},
		{name: "invalid", envValue: "maybe", defaultValue: true, expected: true},
		{name: "empty", envValue: "", defaultValue: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			key := "TEST_BOOL_ENV"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			}

			result := getBoolEnv(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int64
		expected     int64
	}{
		{name: "valid positive", envValue: "123", defaultValue: 0, expected: 123},
		{name: "valid negative", envValue: "-456", defaultValue: 0, expected: -456},
		{name: "zero", envValue: "0", defaultValue: 999, expected: 0},
		{name: "invalid", envValue: "abc", defaultValue: 42, expected: 42},
		{name: "empty", envValue: "", defaultValue: 100, expected: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			key := "TEST_INT_ENV"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			}

			result := getIntEnv(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFloatEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue float64
		expected     float64
	}{
		{name: "valid decimal", envValue: "0.5", defaultValue: 0.0, expected: 0.5},
		{name: "valid integer", envValue: "1", defaultValue: 0.0, expected: 1.0},
		{name: "zero", envValue: "0.0", defaultValue: 0.9, expected: 0.0},
		{name: "invalid", envValue: "abc", defaultValue: 0.3, expected: 0.3},
		{name: "empty", envValue: "", defaultValue: 0.7, expected: 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			key := "TEST_FLOAT_ENV"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			}

			result := getFloatEnv(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Clean environment
	envVars := []string{
		"DIRT_CHAOS_ENABLED",
		"DIRT_CHAOS_SEED",
		"DIRT_LATENCY_GLOBAL_MS",
		"DIRT_LATENCY_PROJECTS_MS",
		"DIRT_LATENCY_INSTANCES_MS",
		"DIRT_LATENCY_METADATA_MS",
		"DIRT_ERRRATE_PROJECTS",
		"DIRT_ERRRATE_PROJECTS_GET",
		"DIRT_ERRRATE_INSTANCES",
		"DIRT_ERRRATE_METADATA",
		"DIRT_ERROR_TYPES",
		"DIRT_ERROR_WEIGHTS",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}

	t.Run("default config", func(t *testing.T) {
		config := loadConfigFromEnv()
		
		assert.False(t, config.Enabled)
		assert.NotZero(t, config.Seed) // Should have a timestamp-based seed
		assert.Nil(t, config.GlobalLatencyRange)
		assert.Nil(t, config.ProjectsLatencyRange)
		assert.Nil(t, config.InstancesLatencyRange)
		assert.Nil(t, config.MetadataLatencyRange)
		assert.Equal(t, 0.0, config.ProjectsErrorRate)
		assert.Equal(t, 0.0, config.ProjectsGetErrorRate)
		assert.Equal(t, 0.0, config.InstancesErrorRate)
		assert.Equal(t, 0.0, config.MetadataErrorRate)
		assert.Equal(t, []int{503, 500, 429}, config.ErrorTypes)
		assert.Equal(t, []int{3, 2, 1}, config.ErrorWeights)
	})

	t.Run("full config from env", func(t *testing.T) {
		// Set environment variables
		os.Setenv("DIRT_CHAOS_ENABLED", "true")
		os.Setenv("DIRT_CHAOS_SEED", "12345")
		os.Setenv("DIRT_LATENCY_GLOBAL_MS", "10-100")
		os.Setenv("DIRT_LATENCY_PROJECTS_MS", "5-50")
		os.Setenv("DIRT_LATENCY_INSTANCES_MS", "20-200")
		os.Setenv("DIRT_LATENCY_METADATA_MS", "1-10")
		os.Setenv("DIRT_ERRRATE_PROJECTS", "0.1")
		os.Setenv("DIRT_ERRRATE_PROJECTS_GET", "0.05")
		os.Setenv("DIRT_ERRRATE_INSTANCES", "0.2")
		os.Setenv("DIRT_ERRRATE_METADATA", "0.15")
		os.Setenv("DIRT_ERROR_TYPES", "500,503")
		os.Setenv("DIRT_ERROR_WEIGHTS", "5,3")

		defer func() {
			for _, v := range envVars {
				os.Unsetenv(v)
			}
		}()

		config := loadConfigFromEnv()

		assert.True(t, config.Enabled)
		assert.Equal(t, int64(12345), config.Seed)
		
		assert.NotNil(t, config.GlobalLatencyRange)
		assert.Equal(t, 10, config.GlobalLatencyRange.Min)
		assert.Equal(t, 100, config.GlobalLatencyRange.Max)
		
		assert.NotNil(t, config.ProjectsLatencyRange)
		assert.Equal(t, 5, config.ProjectsLatencyRange.Min)
		assert.Equal(t, 50, config.ProjectsLatencyRange.Max)
		
		assert.Equal(t, 0.1, config.ProjectsErrorRate)
		assert.Equal(t, 0.05, config.ProjectsGetErrorRate)
		assert.Equal(t, 0.2, config.InstancesErrorRate)
		assert.Equal(t, 0.15, config.MetadataErrorRate)
		
		assert.Equal(t, []int{500, 503}, config.ErrorTypes)
		assert.Equal(t, []int{5, 3}, config.ErrorWeights)
	})
}

func TestChaosService_selectWeightedErrorType(t *testing.T) {
	tests := []struct {
		name         string
		errorTypes   []int
		errorWeights []int
		iterations   int
		expectAll    []int
	}{
		{
			name:         "single error type",
			errorTypes:   []int{500},
			errorWeights: []int{1},
			iterations:   10,
			expectAll:    []int{500},
		},
		{
			name:         "no weights",
			errorTypes:   []int{500, 503},
			errorWeights: []int{},
			iterations:   100,
			expectAll:    []int{500, 503}, // Should use uniform distribution
		},
		{
			name:         "mismatched weights",
			errorTypes:   []int{500, 503},
			errorWeights: []int{1},
			iterations:   100,
			expectAll:    []int{500, 503}, // Should use uniform distribution
		},
		{
			name:         "zero weights",
			errorTypes:   []int{500},
			errorWeights: []int{0},
			iterations:   10,
			expectAll:    []int{500}, // Should return first type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ErrorTypes:   tt.errorTypes,
				ErrorWeights: tt.errorWeights,
			}
			
			service := &ChaosService{
				config: config,
				rng:    rand.New(rand.NewSource(42)), // Use fixed seed for deterministic tests
			}

			seen := make(map[int]bool)
			for i := 0; i < tt.iterations; i++ {
				errorType := service.selectWeightedErrorType()
				seen[errorType] = true
			}

			// Verify we only see expected error types
			for errorType := range seen {
				assert.Contains(t, tt.expectAll, errorType)
			}
		})
	}
}

func TestChaosService_maybeInjectError(t *testing.T) {
	tests := []struct {
		name        string
		errorRate   float64
		expectError bool
	}{
		{
			name:        "zero error rate",
			errorRate:   0.0,
			expectError: false,
		},
		{
			name:        "negative error rate",
			errorRate:   -0.1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ErrorTypes:   []int{500},
				ErrorWeights: []int{1},
			}
			
			service := &ChaosService{
				config: config,
				rng:    rand.New(rand.NewSource(42)), // Use fixed seed
			}

			err := service.maybeInjectError(tt.errorRate)
			
			if tt.expectError {
				assert.Error(t, err)
				dirtErr, ok := err.(*domain.DirtError)
				assert.True(t, ok)
				assert.Contains(t, []string{domain.ErrorCodeInternalError, domain.ErrorCodeServiceUnavailable, domain.ErrorCodeTooManyRequests}, dirtErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Test 100% error rate separately
	t.Run("100% error rate", func(t *testing.T) {
		config := &Config{
			ErrorTypes:   []int{500},
			ErrorWeights: []int{1},
		}
		
		service := &ChaosService{
			config: config,
			rng:    rand.New(rand.NewSource(42)),
		}

		// Test multiple times to ensure we get errors
		errorCount := 0
		iterations := 100
		for i := 0; i < iterations; i++ {
			err := service.maybeInjectError(1.0)
			if err != nil {
				errorCount++
			}
		}
		
		// With 100% error rate, we should get some errors
		assert.Greater(t, errorCount, iterations/2, "Expected more than half errors with 100% rate")
	})
}

func TestChaosService_ApplyChaos_Disabled(t *testing.T) {
	config := &Config{
		Enabled: false,
	}
	
	service := &ChaosService{
		config: config,
		rng:    nil, // Should not be used when disabled
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	ctx := context.Background()

	// Should not apply any chaos when disabled
	err := service.ApplyProjectsChaos(ctx, req, "GET")
	assert.NoError(t, err)

	err = service.ApplyInstancesChaos(ctx, req)
	assert.NoError(t, err)

	err = service.ApplyMetadataChaos(ctx, req)
	assert.NoError(t, err)
}

func TestChaosService_ApplyChaos_BypassHeader(t *testing.T) {
	config := &Config{
		Enabled:           true,
		ProjectsErrorRate: 0.1, // 10% error rate
	}
	
	service := &ChaosService{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Dirt-No-Chaos", "true")
	ctx := context.Background()

	// Should bypass chaos with header
	err := service.ApplyProjectsChaos(ctx, req, "GET")
	assert.NoError(t, err)
}

func TestChaosService_ApplyChaos_ForcedLatency(t *testing.T) {
	config := &Config{
		Enabled: true,
	}
	
	service := &ChaosService{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Dirt-Latency", "10") // Force 10ms latency
	
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := service.ApplyProjectsChaos(ctx, req, "GET")
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.True(t, duration >= 10*time.Millisecond, "Expected at least 10ms delay, got %v", duration)
}

