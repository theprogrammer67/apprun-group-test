package apprun_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	apprun "run-group-test/pkg/app_run"
)

func TestZero(t *testing.T) {
	var g apprun.Group
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestOne(t *testing.T) {
	myError := errors.New("foobar")
	var g apprun.Group
	g.Add(func() error { return myError }, func(error) {})
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := myError, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestMany(t *testing.T) {
	interrupt := errors.New("interrupt")
	var g apprun.Group
	g.Add(func() error { return interrupt }, func(error) {})
	cancel := make(chan struct{})
	g.Add(func() error { <-cancel; return nil }, func(error) { close(cancel) })
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := interrupt, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("timeout")
	}
}

func TestAddAfter(t *testing.T) {
	const count = 10
	var have [count]int
	num := make(chan int, count)

	var g apprun.Group
	var prev *apprun.StartSignal
	for i := 0; i < len(have); i++ {
		n := i
		prev = g.AddAfter(
			func(started *apprun.StartSignal) error {
				num <- n
				started.Success()
				return nil
			},
			func(error) {},
			prev,
		)
	}

	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}

		var want [count]int
		for i := 0; i < len(have); i++ {
			want[i] = i
			select {
			case have[i] = <-num:
			case <-time.After(100 * time.Millisecond):
				t.Errorf("timeout")
			}
		}

		if want != have {
			t.Errorf("incorrect execution order: want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestAddAfterError(t *testing.T) {
	whant := errors.New("test error")

	var g apprun.Group
	first := g.AddAfter(
		func(started *apprun.StartSignal) error {
			started.Error(whant)
			return whant
		},
		func(error) {},
		nil,
	)

	g.AddAfter(
		func(started *apprun.StartSignal) error {
			t.Error("goroutine started although the previous goroutine exited with an error")

			return nil
		},
		func(error) {},
		first,
	)

	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if !errors.Is(err, whant) {
			t.Errorf("incorrect error: %v", err)
		}
	case <-time.After(100 * time.Hour):
		t.Error("timeout")
	}

}

func TestAfterStart(t *testing.T) {
	var start, afterStart bool
	var mx sync.Mutex

	var g apprun.Group
	ss := g.AddAfter(
		func(started *apprun.StartSignal) error {
			mx.Lock()
			start = true
			mx.Unlock()

			started.Success()

			return nil
		},
		func(error) {},
		nil,
	)

	ss.AferStart = func() {
		mx.Lock()
		started := start
		mx.Unlock()

		if !started {
			t.Error("goroutine did not start before AferStart func")
		}
		afterStart = true
	}

	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}

		if !start {
			t.Error("goroutine did not start")
		}
		if !afterStart {
			t.Error("function AferStart not executed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestRunApp(t *testing.T) {
	// TO DO
}
