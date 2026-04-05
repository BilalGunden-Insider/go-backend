package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failCount    int
	maxFailures  int
	resetTimeout time.Duration
	lastFailTime time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        Closed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case Open:
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = HalfOpen
			cb.mu.Unlock()
			return cb.tryCall(fn)
		}
		cb.mu.Unlock()
		return ErrCircuitOpen

	case HalfOpen:
		cb.mu.Unlock()
		return cb.tryCall(fn)

	default:
		cb.mu.Unlock()
		return cb.tryCall(fn)
	}
}

func (cb *CircuitBreaker) tryCall(fn func() error) error {
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failCount++
		cb.lastFailTime = time.Now()
		if cb.failCount >= cb.maxFailures {
			cb.state = Open
		}
		return err
	}

	cb.failCount = 0
	cb.state = Closed
	return nil
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
