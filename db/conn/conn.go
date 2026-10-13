// Package conn holds the client for one backing system (a Postgres pool, a
// Redis client, ...) and owns its lifecycle, so each backend package only
// says how to open and close its client.
package conn

import (
	"fmt"
	"sync"

	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l = log.NewLogger("conn")

// Conn holds one client of type T.
//
// Get opens the client with the default opener on first use. Init opens one
// with explicit options and replaces whatever was open. Conn never pings or
// re-opens a client that is open: every client it holds reconnects on its own.
type Conn[T any] struct {
	name        string
	openDefault func() (T, error)
	close       func(T) error

	mu     sync.Mutex
	client T
	open   bool
}

// New returns a Conn that opens with openDefault on first use and closes
// replaced clients with close.
func New[T any](name string, openDefault func() (T, error), close func(T) error) *Conn[T] {
	return &Conn[T]{name: name, openDefault: openDefault, close: close}
}

// Init opens a client with open and makes it the current one, closing the
// previous client. If open fails, the previous client stays.
func (c *Conn[T]) Init(open func() (T, error)) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.replace(open)
}

// InitDefault opens the client with the default opener unless one is
// already open.
func (c *Conn[T]) InitDefault() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.open {
		return nil
	}
	return c.replace(c.openDefault)
}

// Get returns the current client, opening it with the default opener if
// none is open.
func (c *Conn[T]) Get() (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.open {
		if err := c.replace(c.openDefault); err != nil {
			var zero T
			return zero, err
		}
	}
	return c.client, nil
}

// MustGet is Get, panicking if the client can't be opened.
func (c *Conn[T]) MustGet() T {
	client, err := c.Get()
	if err != nil {
		panic(err)
	}
	return client
}

// Close closes the current client, if any. The next Get opens a new one.
func (c *Conn[T]) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.open {
		return nil
	}
	err := c.close(c.client)
	var zero T
	c.client, c.open = zero, false
	return err
}

// replace must be called with c.mu held.
func (c *Conn[T]) replace(open func() (T, error)) error {
	client, err := open()
	if err != nil {
		l.Error("failed to open", zap.String("conn", c.name), zap.Error(err))
		return fmt.Errorf("%s: %w", c.name, err)
	}
	if c.open {
		if err := c.close(c.client); err != nil {
			l.Warn("failed to close replaced client", zap.String("conn", c.name), zap.Error(err))
		}
	}
	c.client, c.open = client, true
	l.Info("opened", zap.String("conn", c.name))
	return nil
}
