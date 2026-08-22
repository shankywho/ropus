package resilience

import (
	"errors"
	"sync"
	"time"
)

// State defines the Circuit Breaker lifecycle states.
type State string

const (
	StateClosed   State = "CLOSED"
	StateHalfOpen State = "HALF_OPEN"
	StateOpen     State = "OPEN"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open; fast failing request")
)

// CircuitBreakerConfig holds threshold parameters.
type CircuitBreakerConfig struct {
	Name             string
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

// CircuitBreaker provides resilient dependency wrapping.
type CircuitBreaker struct {
	mu           sync.RWMutex
	name         string
	state        State
	failures     int
	successes    int
	threshold    int
	reqThreshold int
	timeout      time.Duration
	lastStateChange time.Time
}

// NewCircuitBreaker initializes a circuit breaker.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &CircuitBreaker{
		name:            cfg.Name,
		state:           StateClosed,
		threshold:       cfg.FailureThreshold,
		reqThreshold:    cfg.SuccessThreshold,
		timeout:         cfg.Timeout,
		lastStateChange: time.Now(),
	}
}

// Execute runs the wrapped function if the circuit permits.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if Open timeout has expired -> transition to Half-Open
	if cb.state == StateOpen && time.Since(cb.lastStateChange) > cb.timeout {
		cb.state = StateHalfOpen
		cb.successes = 0
		cb.lastStateChange = time.Now()
	}

	if cb.state == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}

	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		if cb.state == StateHalfOpen || cb.failures >= cb.threshold {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
		return err
	}

	// Success handling
	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.reqThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.lastStateChange = time.Now()
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}

	return nil
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
