package chaos

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nicolas/dirtcloud/domain"
)

// Config holds chaos engineering configuration
type Config struct {
	Enabled bool
	Seed    int64
	
	// Global latency range in milliseconds
	GlobalLatencyRange *LatencyRange
	
	// Per-resource latency ranges
	ProjectsLatencyRange  *LatencyRange
	InstancesLatencyRange *LatencyRange
	MetadataLatencyRange  *LatencyRange
	
	// Per-resource error rates
	ProjectsErrorRate    float64
	ProjectsGetErrorRate float64
	InstancesErrorRate   float64
	MetadataErrorRate    float64
	
	// Error configuration
	ErrorTypes   []int
	ErrorWeights []int
}

// LatencyRange defines min-max latency in milliseconds
type LatencyRange struct {
	Min int
	Max int
}

// ChaosService provides chaos engineering capabilities
type ChaosService struct {
	config *Config
	rng    *rand.Rand
}

// NewChaosService creates a new chaos service from environment variables
func NewChaosService() *ChaosService {
	config := loadConfigFromEnv()
	
	var rng *rand.Rand
	if config.Enabled {
		rng = rand.New(rand.NewSource(config.Seed))
	}
	
	return &ChaosService{
		config: config,
		rng:    rng,
	}
}

// loadConfigFromEnv loads chaos configuration from environment variables
func loadConfigFromEnv() *Config {
	config := &Config{
		Enabled: getBoolEnv("DIRT_CHAOS_ENABLED", false),
		Seed:    getIntEnv("DIRT_CHAOS_SEED", time.Now().UnixNano()),
		
		ErrorTypes:   []int{503, 500, 429}, // defaults
		ErrorWeights: []int{3, 2, 1},       // defaults
	}
	
	// Load latency ranges
	if latency := getEnv("DIRT_LATENCY_GLOBAL_MS", ""); latency != "" {
		config.GlobalLatencyRange = parseLatencyRange(latency)
	}
	if latency := getEnv("DIRT_LATENCY_PROJECTS_MS", ""); latency != "" {
		config.ProjectsLatencyRange = parseLatencyRange(latency)
	}
	if latency := getEnv("DIRT_LATENCY_INSTANCES_MS", ""); latency != "" {
		config.InstancesLatencyRange = parseLatencyRange(latency)
	}
	if latency := getEnv("DIRT_LATENCY_METADATA_MS", ""); latency != "" {
		config.MetadataLatencyRange = parseLatencyRange(latency)
	}
	
	// Load error rates
	config.ProjectsErrorRate = getFloatEnv("DIRT_ERRRATE_PROJECTS", 0.0)
	config.ProjectsGetErrorRate = getFloatEnv("DIRT_ERRRATE_PROJECTS_GET", config.ProjectsErrorRate)
	config.InstancesErrorRate = getFloatEnv("DIRT_ERRRATE_INSTANCES", 0.0)
	config.MetadataErrorRate = getFloatEnv("DIRT_ERRRATE_METADATA", 0.0)
	
	// Load error types and weights
	if types := getEnv("DIRT_ERROR_TYPES", ""); types != "" {
		config.ErrorTypes = parseIntList(types)
	}
	if weights := getEnv("DIRT_ERROR_WEIGHTS", ""); weights != "" {
		config.ErrorWeights = parseIntList(weights)
	}
	
	return config
}

// ApplyProjectsChaos applies chaos to projects operations
func (c *ChaosService) ApplyProjectsChaos(ctx context.Context, r *http.Request, method string) error {
	if !c.config.Enabled {
		return nil
	}
	
	// Check for bypass header
	if r.Header.Get("X-Dirt-No-Chaos") == "true" {
		return nil
	}
	
	// Apply latency
	c.applyLatency(ctx, r, c.config.ProjectsLatencyRange)
	
	// Apply error injection
	errorRate := c.config.ProjectsErrorRate
	if method == "GET" && c.config.ProjectsGetErrorRate > 0 {
		errorRate = c.config.ProjectsGetErrorRate
	}
	
	return c.maybeInjectError(errorRate)
}

// ApplyInstancesChaos applies chaos to instances operations
func (c *ChaosService) ApplyInstancesChaos(ctx context.Context, r *http.Request) error {
	if !c.config.Enabled {
		return nil
	}
	
	// Check for bypass header
	if r.Header.Get("X-Dirt-No-Chaos") == "true" {
		return nil
	}
	
	// Apply latency
	c.applyLatency(ctx, r, c.config.InstancesLatencyRange)
	
	// Apply error injection
	return c.maybeInjectError(c.config.InstancesErrorRate)
}

// ApplyMetadataChaos applies chaos to metadata operations
func (c *ChaosService) ApplyMetadataChaos(ctx context.Context, r *http.Request) error {
	if !c.config.Enabled {
		return nil
	}
	
	// Check for bypass header
	if r.Header.Get("X-Dirt-No-Chaos") == "true" {
		return nil
	}
	
	// Apply latency
	c.applyLatency(ctx, r, c.config.MetadataLatencyRange)
	
	// Apply error injection
	return c.maybeInjectError(c.config.MetadataErrorRate)
}

// applyLatency applies latency injection
func (c *ChaosService) applyLatency(ctx context.Context, r *http.Request, resourceRange *LatencyRange) {
	// Check for forced latency header
	if forcedLatency := r.Header.Get("X-Dirt-Latency"); forcedLatency != "" {
		if ms, err := strconv.Atoi(forcedLatency); err == nil && ms > 0 {
			select {
			case <-time.After(time.Duration(ms) * time.Millisecond):
			case <-ctx.Done():
			}
			return
		}
	}
	
	// Determine latency range to use (resource-specific overrides global)
	latencyRange := c.config.GlobalLatencyRange
	if resourceRange != nil {
		latencyRange = resourceRange
	}
	
	if latencyRange == nil {
		return
	}
	
	// Calculate random latency within range
	latency := latencyRange.Min
	if latencyRange.Max > latencyRange.Min {
		latency += c.rng.Intn(latencyRange.Max - latencyRange.Min + 1)
	}
	
	if latency > 0 {
		select {
		case <-time.After(time.Duration(latency) * time.Millisecond):
		case <-ctx.Done():
		}
	}
}

// maybeInjectError randomly injects an error based on error rate
func (c *ChaosService) maybeInjectError(errorRate float64) error {
	if errorRate <= 0.0 || c.rng.Float64() > errorRate {
		return nil
	}
	
	// Select error type based on weights
	errorCode := c.selectWeightedErrorType()
	
	switch errorCode {
	case 429:
		return domain.TooManyRequestsError("chaos: rate limited")
	case 500:
		return domain.InternalError("chaos: internal server error")
	case 503:
		return domain.ServiceUnavailableError("chaos: service unavailable")
	default:
		return domain.InternalError("chaos: unknown error")
	}
}

// selectWeightedErrorType selects an error type based on configured weights
func (c *ChaosService) selectWeightedErrorType() int {
	if len(c.config.ErrorTypes) == 0 {
		return 500
	}
	
	if len(c.config.ErrorWeights) != len(c.config.ErrorTypes) {
		// If weights don't match types, use uniform distribution
		return c.config.ErrorTypes[c.rng.Intn(len(c.config.ErrorTypes))]
	}
	
	// Calculate total weight
	totalWeight := 0
	for _, weight := range c.config.ErrorWeights {
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return c.config.ErrorTypes[0]
	}
	
	// Select based on weights
	target := c.rng.Intn(totalWeight)
	currentWeight := 0
	
	for i, weight := range c.config.ErrorWeights {
		currentWeight += weight
		if target < currentWeight {
			return c.config.ErrorTypes[i]
		}
	}
	
	// Fallback
	return c.config.ErrorTypes[0]
}

// Utility functions for parsing environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func getIntEnv(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intValue
	}
	
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	
	return defaultValue
}

func parseLatencyRange(value string) *LatencyRange {
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return nil
	}
	
	min, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	max, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	
	if err1 != nil || err2 != nil || min < 0 || max < min {
		return nil
	}
	
	return &LatencyRange{Min: min, Max: max}
}

func parseIntList(value string) []int {
	parts := strings.Split(value, ",")
	var result []int
	
	for _, part := range parts {
		if intValue, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result = append(result, intValue)
		}
	}
	
	return result
}