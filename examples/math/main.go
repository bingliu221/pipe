package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/bingliu221/pipe"
)

func gen(cancel context.CancelFunc) func() interface{} {
	value := 0

	return func() interface{} {
		value++
		if value >= 10 {
			cancel()
		}
		return value
	}
}

func add(n int) func(interface{}) interface{} {
	return func(x interface{}) interface{} {
		return x.(int) + n
	}
}

func multiply(n int) func(interface{}) interface{} {
	return func(x interface{}) interface{} {
		return x.(int) * n
	}
}

func print() func(interface{}) {
	return func(x interface{}) {
		fmt.Println(x.(int))
	}
}

func main() {

	mainCtx, cancel := context.WithCancel(context.Background())

	p := pipe.Start(gen(cancel), mainCtx).Then(add(2)).Then(multiply(2)).End(print())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	for {
		select {
		case sig := <-signalChan:
			if sig == os.Interrupt {
				p.Cancel()
			}
		case <-p.Done():
			return
		}
	}
}
