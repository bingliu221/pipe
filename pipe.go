package pipe

import (
	"context"
)

type Pipe struct {
	Done   func() <-chan struct{}
	Cancel context.CancelFunc

	C chan interface{}
}

func Start(f func() interface{}) Pipe {
	p := Pipe{}

	currentCtx, cancel := context.WithCancel(context.Background())
	currentChan := make(chan interface{})

	go func() {
		defer cancel()
		defer close(currentChan)

	ForLoop:
		for {
			select {
			case <-currentCtx.Done():
				break ForLoop
			default:
				out := f()
				currentChan <- out
			}
		}
	}()

	p.Done = currentCtx.Done
	p.C = currentChan
	p.Cancel = cancel
	return p
}

func (p Pipe) Then(f func(interface{}) interface{}) Pipe {
	frontChan := p.C

	currentCtx, cancel := context.WithCancel(context.Background())
	currentChan := make(chan interface{})

	go func() {
		defer cancel()
		defer close(currentChan)

	ForLoop:
		for {
			select {
			case <-currentCtx.Done():
				break ForLoop
			case in, ok := <-frontChan:
				if ok {
					out := f(in)
					currentChan <- out
				} else {
					break ForLoop
				}
			}
		}
	}()

	p.Done = currentCtx.Done
	p.C = currentChan
	return p
}

func (p Pipe) End(f func(interface{})) Pipe {
	frontChan := p.C

	currentCtx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

	ForLoop:
		for {
			select {
			case <-currentCtx.Done():
				break ForLoop
			case in, ok := <-frontChan:
				if ok {
					f(in)
				} else {
					break ForLoop
				}
			}
		}
	}()

	p.Done = currentCtx.Done
	p.C = nil
	return p
}
