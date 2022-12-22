package gc

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/filters"
	ci "github.com/taubyte/go-simple-containers"
)

type config struct {
	interval time.Duration
	maxAge   time.Duration
	filters  filters.Args
}

type Option func(o *config) error

func Interval(t time.Duration) Option {
	return func(o *config) error {
		o.interval = t
		return nil
	}
}

func MaxAge(t time.Duration) Option {
	return func(o *config) error {
		o.maxAge = t
		return nil
	}
}

func Filter(key, value string) Option {
	return func(o *config) error {
		o.filters.Add(key, value)
		return nil
	}
}

// Starts a new garbage collector with the specified interval check, and removes containers older than specified age.
func Start(ctx context.Context, options ...Option) error {
	client, err := ci.New()
	if err != nil {
		return err
	}

	cnf := &config{}
	for _, opt := range options {
		if err := opt(cnf); err != nil {
			return err
		}
	}

	go func() {
		defer client.Close()
		for {
			select {
			case <-time.After(cnf.interval):
				client.Clean(ctx, cnf.maxAge, cnf.filters)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
