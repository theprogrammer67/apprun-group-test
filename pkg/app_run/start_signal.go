package apprun

import "context"

type StartSignal struct {
	ch            chan bool
	AferStartFunc func()
}

func NewStartSignal() *StartSignal {
	return &StartSignal{
		ch: make(chan bool, 1),
	}
}

func (s *StartSignal) Success() {
	s.ch <- true
	if s.AferStartFunc != nil {
		s.AferStartFunc()
	}
}

func (s *StartSignal) Error() {
	s.ch <- false
}

func (s *StartSignal) Wait() bool {
	return <-s.ch
}

func (s *StartSignal) WaitWithContext(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case res := <-s.ch:
		return res
	}
}
