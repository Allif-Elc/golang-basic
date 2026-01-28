package concurrency

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ConcurrentProcessor provides bounded concurrent processing using errgroup.
// Per Specs: always bound concurrency to prevent unbounded goroutine spawns.
type ConcurrentProcessor struct {
	maxConcurrency int
}

// NewConcurrentProcessor creates a new processor with bounded concurrency.
// maxConcurrency should be set based on the operation type (CPU-bound vs I/O-bound).
// For CPU-bound: use runtime.NumCPU()
// For I/O-bound: use runtime.NumCPU() * 2-4
func NewConcurrentProcessor(maxWorkers int) *ConcurrentProcessor {
	return &ConcurrentProcessor{maxConcurrency: maxWorkers}
}

// Process executes the provided function concurrently for each item,
// with bounded concurrency controlled by errgroup.SetLimit.
// Per Specs: capture loop variable to avoid closure issues.
func (p *ConcurrentProcessor) Process(ctx context.Context, items []interface{}, fn func(context.Context, interface{}) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(p.maxConcurrency)

	for _, item := range items {
		item := item // Capture loop variable (specs requirement)
		g.Go(func() error {
			return fn(ctx, item)
		})
	}
	return g.Wait()
}

// ProcessTyped is a type-safe wrapper for Process using generics.
// T is the type of items to process.
func ProcessTyped[T any](ctx context.Context, items []T, maxWorkers int, fn func(context.Context, T) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxWorkers)

	for _, item := range items {
		item := item // Capture loop variable
		g.Go(func() error {
			return fn(ctx, item)
		})
	}
	return g.Wait()
}
