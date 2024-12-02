package apprun_test

import (
	"errors"
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

	g.Run()

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
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestAfterStart(t *testing.T) {
}
