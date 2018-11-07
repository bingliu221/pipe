package pipe

import "context"

type Pipe struct {
	Ctx      context.Context
	lastChan chan interface{}

	Cancel context.CancelFunc
}

func Start(f func() interface{}) Pipe {
	p := Pipe{}
	startCtx, cancel := context.WithCancel(context.Background())
	p.Cancel = cancel

	currentCtx, cancel := context.WithCancel(context.Background())

	currentChan := make(chan interface{})

	go func() {
		for {
			select {
			case <-startCtx.Done():
				close(currentChan)
				cancel()
				return
			default:
				out := f()
				currentChan <- out
			}
		}
	}()

	p.Ctx = currentCtx
	p.lastChan = currentChan
	p.Cancel = cancel
	return p
}

func (p Pipe) Then(f func(interface{}) interface{}) Pipe {
	currentCtx, cancel := context.WithCancel(context.Background())
	lastCtx := p.Ctx
	frontChan := p.lastChan

	currentChan := make(chan interface{})

	go func() {
		for {
			select {
			case <-lastCtx.Done():
				close(currentChan)
				cancel()
				return
			case in := <-frontChan:
				out := f(in)
				currentChan <- out
			}
		}
	}()

	p.Ctx = currentCtx
	p.lastChan = currentChan
	return p
}

func (p Pipe) End(f func(interface{})) Pipe {
	currentCtx, cancel := context.WithCancel(context.Background())
	lastCtx := p.Ctx
	frontChan := p.lastChan

	go func() {
		for {
			select {
			case <-lastCtx.Done():
				cancel()
				return
			case in := <-frontChan:
				f(in)
			}
		}
	}()

	p.Ctx = currentCtx
	p.lastChan = nil
	return p
}
