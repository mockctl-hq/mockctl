package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mockctl-hq/mockctl/internal/shared"
)

// StandardChaosEvaluator implements shared.ChaosEvaluator.
type StandardChaosEvaluator struct {
	mu         sync.RWMutex
	enabled    bool
	errorRate  int
	latencyMs  int
	statusCode int
	clock      shared.Clock
}

var _ shared.ChaosEvaluator = (*StandardChaosEvaluator)(nil)

// NewStandardChaosEvaluator initializes a new chaos evaluator.
func NewStandardChaosEvaluator(clock shared.Clock) *StandardChaosEvaluator {
	return &StandardChaosEvaluator{
		clock:      clock,
		statusCode: 500,
	}
}

// UpdateConfig updates the chaos rules dynamically with default 500 status.
func (c *StandardChaosEvaluator) UpdateConfig(ctx context.Context, errorRate int, latencyMs int) {
	c.UpdateConfigWithStatus(ctx, errorRate > 0 || latencyMs > 0, errorRate, latencyMs, 500)
}

// UpdateConfigWithStatus updates chaos rules dynamically including enabled flag and custom status code.
func (c *StandardChaosEvaluator) UpdateConfigWithStatus(ctx context.Context, enabled bool, errorRate int, latencyMs int, statusCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
	c.errorRate = errorRate
	c.latencyMs = latencyMs
	if statusCode <= 0 {
		statusCode = 500
	}
	c.statusCode = statusCode
}

// Config returns the active chaos settings (enabled, errorRate, latencyMs, statusCode).
func (c *StandardChaosEvaluator) Config() (bool, int, int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	status := c.statusCode
	if status <= 0 {
		status = 500
	}
	return c.enabled, c.errorRate, c.latencyMs, status
}

// Evaluate applies latency and returns an error status code if chaos strikes.
func (c *StandardChaosEvaluator) Evaluate(ctx context.Context) (int, error) {
	c.mu.RLock()
	enabled := c.enabled
	latency := c.latencyMs
	errRate := c.errorRate
	status := c.statusCode
	if status <= 0 {
		status = 500
	}
	c.mu.RUnlock()

	if !enabled {
		return 0, nil
	}

	if latency > 0 {
		select {
		case <-c.clock.After(time.Duration(latency) * time.Millisecond):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	if errRate > 0 {
		// #nosec G404 -- non-cryptographic RNG is perfectly fine for simulated chaos
		if rand.Intn(100) < errRate {
			return status, fmt.Errorf("simulated chaos failure (HTTP %d)", status)
		}
	}

	return 0, nil
}
