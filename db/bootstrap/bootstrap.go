// Package bootstrap prepares a database from SQL files and statements.
// Run owns discovery, ordering, cancellation and failure policy; adapters only
// select a connection and execute SQL. It never begins a transaction or closes
// the connection, so earlier changes may remain when a later step fails.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Plan runs explicit Files, then each directory's sorted *.sql matches, then
// Tables and Indexes, preserving order and duplicates. Each file is executed
// whole, without splitting SQL on semicolons.
type Plan struct {
	Files   []string
	Dirs    []string
	Tables  []string
	Indexes []string
}

// Execute executes SQL on an already selected connection using ctx.
type Execute func(ctx context.Context, sql string) error

// Open selects a connection without transferring ownership to Run.
type Open func(ctx context.Context) (Execute, error)

// Run discovers the plan's files, selects a connection once if there is work,
// and stops at the first failure or cancellation. An empty plan does not open a
// connection. Step labels omit SQL and credentials; underlying errors are
// preserved for errors.Is and errors.As.
func Run(ctx context.Context, plan Plan, open Open) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	files := append([]string(nil), plan.Files...)
	for _, dir := range plan.Dirs {
		if err := ctx.Err(); err != nil {
			return err
		}
		matches, err := filepath.Glob(filepath.Join(dir, "*.sql"))
		if err != nil {
			return fmt.Errorf("discover SQL directory %q: %w", dir, err)
		}
		files = append(files, matches...)
	}
	if len(files)+len(plan.Tables)+len(plan.Indexes) == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	exec, err := open(ctx)
	if err != nil {
		return fmt.Errorf("open SQL bootstrap connection: %w", err)
	}
	run := func(label, sql string) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
		if err := exec(ctx, sql); err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil && !errors.Is(err, ctxErr) {
				err = errors.Join(err, ctxErr)
			}
			return fmt.Errorf("%s: %w", label, err)
		}
		return nil
	}
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("SQL file %q: %w", file, err)
		}
		b, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read SQL file %q: %w", file, err)
		}
		if err := run(fmt.Sprintf("execute SQL file %q", file), string(b)); err != nil {
			return err
		}
	}
	for i, stmt := range plan.Tables {
		if err := run(fmt.Sprintf("execute table statement %d", i+1), stmt); err != nil {
			return err
		}
	}
	for i, stmt := range plan.Indexes {
		if err := run(fmt.Sprintf("execute index statement %d", i+1), stmt); err != nil {
			return err
		}
	}
	return ctx.Err()
}
