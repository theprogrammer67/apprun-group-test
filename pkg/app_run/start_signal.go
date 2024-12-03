package apprun

type StartSignal struct {
	ch        chan error
	AferStart func()
}

func NewStartSignal() *StartSignal {
	return &StartSignal{
		ch: make(chan error, 1),
	}
}

func (s *StartSignal) Success() {
	if s.AferStart != nil {
		s.AferStart()
	}
	s.ch <- nil
}

func (s *StartSignal) Error(err error) {
	s.ch <- err
}

func (s *StartSignal) Wait() error {
	return <-s.ch
}
