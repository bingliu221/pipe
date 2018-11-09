package pipe

import "context"

type Pipe struct {
	ctx      context.Context
	lastChan chan interface{}

	Cancel context.CancelFunc
}

func Start(f func() interface{}, parent context.Context) Pipe {
	p := Pipe{}
	if parent == nil {
		parent = context.Background()
	}

	currentCtx, cancel := context.WithCancel(context.Background())

	currentChan := make(chan interface{})

	go func() {
		defer cancel()
		for {
			select {
			case <-parent.Done():
				close(currentChan)
				return
			default:
				out := f()
				currentChan <- out
			}
		}
	}()

	p.ctx = currentCtx
	p.lastChan = currentChan
	p.Cancel = cancel
	return p
}

func (p Pipe) Then(f func(interface{}) interface{}) Pipe {
	currentCtx, cancel := context.WithCancel(context.Background())
	lastCtx := p.ctx
	frontChan := p.lastChan

	currentChan := make(chan interface{})

	go func() {
		defer cancel()
		for {
			select {
			case <-lastCtx.Done():
				close(currentChan)
				return
			case in, ok := <-frontChan:
				if ok {
					out := f(in)
					currentChan <- out
				}
			}
		}
	}()

	p.ctx = currentCtx
	p.lastChan = currentChan
	return p
}

func (p Pipe) End(f func(interface{})) Pipe {
	currentCtx, cancel := context.WithCancel(context.Background())
	lastCtx := p.ctx
	frontChan := p.lastChan

	go func() {
		defer cancel()
		for {
			select {
			case <-lastCtx.Done():
				cancel()
				return
			case in, ok := <-frontChan:
				if ok {
					f(in)
				}
			}
		}
	}()

	p.ctx = currentCtx
	p.lastChan = nil
	return p
}

func (p Pipe) Done() <-chan struct{} {
	return p.ctx.Done()
}
