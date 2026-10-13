// Package logs multiplexes log streams from many services into one channel.
package logs

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

type Line struct{ Service, Text string }

type Options struct {
	Since  time.Duration
	Follow bool
	Tail   string // "all" or a number
}

// Source is the abstraction the Docker client satisfies. Tests use a fake.
type Source interface {
	Stream(ctx context.Context, service string, o Options, out chan<- Line) error
}

// FanIn starts one goroutine per service, all sending into the shared
// out channel (fan-in). It blocks until every stream ends, then returns
// the joined errors. The caller closes out after FanIn returns.
func FanIn(ctx context.Context, src Source, services []string, o Options, out chan<- Line) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(services)) // buffered: goroutines never block on error
	for _, s := range services {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			if err := src.Stream(ctx, s, o, out); err != nil && !errors.Is(err, context.Canceled) {
				errs <- fmt.Errorf("%s: %w", s, err)
			}
		}(s)
	}
	wg.Wait()
	close(errs)

	var all []error
	for e := range errs {
		all = append(all, e)
	}
	return errors.Join(all...)
}

// NewMatcher compiles a regex filter. Empty pattern matches everything.
func NewMatcher(pattern string) (func(string) bool, error) {
	if pattern == "" {
		return func(string) bool { return true }, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return re.MatchString, nil
}
