package utils

import (
	"context"
	"sync"
	"time"
)

// SafeCounter is a thread-safe counter
type SafeCounter struct {
	mu    sync.RWMutex
	value int64
}

// NewSafeCounter creates a new SafeCounter
func NewSafeCounter() *SafeCounter {
	return &SafeCounter{}
}

// Increment increments the counter by 1
func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Add adds the specified value to the counter
func (c *SafeCounter) Add(delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += delta
}

// Get returns the current value of the counter
func (c *SafeCounter) Get() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Reset resets the counter to 0
func (c *SafeCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}

// WaitGroup is a wrapper around sync.WaitGroup with context support
type WaitGroup struct {
	wg sync.WaitGroup
}

// NewWaitGroup creates a new WaitGroup
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// Add adds delta to the WaitGroup counter
func (w *WaitGroup) Add(delta int) {
	w.wg.Add(delta)
}

// Done decrements the WaitGroup counter by one
func (w *WaitGroup) Done() {
	w.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero
func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// WaitWithContext waits for the WaitGroup with context cancellation support
func (w *WaitGroup) WaitWithContext(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Semaphore is a counting semaphore implementation
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the specified capacity
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Acquire acquires a permit from the semaphore
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// TryAcquire tries to acquire a permit without blocking
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// AcquireWithContext acquires a permit with context cancellation support
func (s *Semaphore) AcquireWithContext(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AcquireWithTimeout acquires a permit with a timeout
func (s *Semaphore) AcquireWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.AcquireWithContext(ctx) == nil
}

// Release releases a permit back to the semaphore
func (s *Semaphore) Release() {
	<-s.ch
}

// AvailablePermits returns the number of available permits
func (s *Semaphore) AvailablePermits() int {
	return cap(s.ch) - len(s.ch)
}

// UsedPermits returns the number of used permits
func (s *Semaphore) UsedPermits() int {
	return len(s.ch)
}
