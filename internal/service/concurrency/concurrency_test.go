package concurrency

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewConcurrentProcessor(t *testing.T) {
	p := NewConcurrentProcessor(5)
	if p == nil {
		t.Fatal("processor should not be nil")
	}
}

func TestProcess_Success(t *testing.T) {
	p := NewConcurrentProcessor(5)
	items := []interface{}{1, 2, 3, 4, 5}
	var mu sync.Mutex
	var results []int

	err := p.Process(context.Background(), items, func(ctx context.Context, item interface{}) error {
		mu.Lock()
		results = append(results, item.(int))
		mu.Unlock()
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	mu.Lock()
	count := len(results)
	mu.Unlock()
	if count != 5 {
		t.Errorf("expected 5 results, got %d", count)
	}
}

func TestProcess_Error(t *testing.T) {
	p := NewConcurrentProcessor(2)
	items := []interface{}{1, 2, 3}
	sentinel := errors.New("oops")

	err := p.Process(context.Background(), items, func(ctx context.Context, item interface{}) error {
		if item.(int) == 2 {
			return sentinel
		}
		return nil
	})

	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestProcess_ContextCancelled(t *testing.T) {
	p := NewConcurrentProcessor(2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	items := []interface{}{1, 2, 3}
	err := p.Process(ctx, items, func(ctx context.Context, item interface{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	if err == nil {
		t.Error("expected context cancelled error")
	}
}

func TestProcess_EmptyItems(t *testing.T) {
	p := NewConcurrentProcessor(5)
	err := p.Process(context.Background(), []interface{}{}, func(ctx context.Context, item interface{}) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error for empty items: %v", err)
	}
}

func TestProcessTyped_Success(t *testing.T) {
	items := []string{"a", "b", "c"}
	var mu sync.Mutex
	var results []string

	err := ProcessTyped(context.Background(), items, 3, func(ctx context.Context, item string) error {
		mu.Lock()
		results = append(results, item+"!")
		mu.Unlock()
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	mu.Lock()
	count := len(results)
	mu.Unlock()
	if count != 3 {
		t.Errorf("expected 3 results, got %d", count)
	}
}

func TestProcessTyped_BoundedConcurrency(t *testing.T) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	maxWorkers := 4
	active := make(chan struct{}, maxWorkers)
	_ = ProcessTyped(context.Background(), items, maxWorkers, func(ctx context.Context, item int) error {
		select {
		case active <- struct{}{}:
		default:
			t.Error("more than maxWorkers goroutines running concurrently")
		}
		time.Sleep(10 * time.Millisecond)
		<-active
		return nil
	})
}
