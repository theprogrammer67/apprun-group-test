// Package run implements an actor-runner with deterministic teardown. It is
// somewhat similar to package errgroup, except it does not require actor
// goroutines to understand context semantics. This makes it suitable for use in
// more circumstances; for example, goroutines which are handling connections
// from net.Listeners, or scanning input from a closable io.Reader.
package apprun

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"syscall"
)

// Group collects actors (functions) and runs them concurrently.
// When one actor (function) returns, all actors are interrupted.
// The zero value of a Group is useful.
type Group struct {
	actors []actor
	log    *slog.Logger
}

func New(log *slog.Logger) Group {
	return Group{log: log}
}

// Add an actor (function) to the group. Each actor must be pre-emptable by an
// interrupt function. That is, if interrupt is invoked, execute should return.
// Also, it must be safe to call interrupt even after execute has returned.
//
// The first actor (function) to return interrupts all running actors.
// The error is passed to the interrupt functions, and is returned by Run.
func (g *Group) Add(execute func() error, interrupt func(error)) {
	g.actors = append(g.actors, actor{
		func(started *StartSignal) error { return execute() },
		interrupt,
		NewStartSignal(),
		nil,
	})
}

// Run all actors (functions) concurrently.
// When the first actor returns, all others are interrupted.
// Run only returns when all actors have exited.
// Run returns the error returned by the first exiting actor.
func (g *Group) Run() error {
	if len(g.actors) == 0 {
		return nil
	}

	// Run each actor.
	errors := make(chan error, len(g.actors))
	for _, a := range g.actors {
		go func(a actor) {
			errors <- a.exec()
		}(a)
	}

	// Wait for the first actor to stop.
	err := <-errors

	// Signal all actors to stop.
	for _, a := range g.actors {
		a.interrupt(err)
	}

	// Wait for all actors to stop.
	for i := 1; i < cap(errors); i++ {
		<-errors
	}

	// Return the original error.
	return err
}

// Adds the SignalHandler actor and run all actors
func (g *Group) RunApp(ctx context.Context) error {
	if len(g.actors) == 0 {
		return nil
	}

	g.addSignalHandler(ctx)

	return g.Run()
}

// Add an actor (function) to the group.
// Returns a signal that allows notifying another actor of a successful run.
// The run of an actor may depend on a signal from another actor.
func (g *Group) AddAfter(execute func(started *StartSignal) error, interrupt func(error), after *StartSignal) *StartSignal {

	actor := actor{
		execute:   execute,
		interrupt: interrupt,
		started:   NewStartSignal(),
		after:     after,
	}
	g.actors = append(g.actors, actor)

	return actor.started
}

// Adds the SignalHandler actor
func (g *Group) addSignalHandler(ctx context.Context) {
	execute, interrupt := SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM)
	g.Add(func() error {
		err := execute()
		if errors.As(err, &SignalError{}) {
			g.log.Warn(err.Error())

			return nil
		}

		return err
	}, func(err error) {
		interrupt(err)
	})
}

type actor struct {
	execute   func(started *StartSignal) error
	interrupt func(error)
	started   *StartSignal
	after     *StartSignal
}

func (a *actor) exec() error {
	if a.after != nil {
		err := a.after.Wait()
		if err != nil {
			err := fmt.Errorf("error starting previous group item: %w", err)
			a.started.Error(err)
			return err
		}
	}

	return a.execute(a.started)
}
