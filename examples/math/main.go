package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bingliu221/pipe"
)

func gen(cancel []context.CancelFunc) func() interface{} {
	value := 0

	return func() interface{} {
		value++
		if value >= 10 {
			(cancel[0])()
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
		log.Println(x.(int))
	}
}

func main() {

	cancel := make([]context.CancelFunc, 1)

	p := pipe.Start(gen(cancel)).Then(add(2)).Then(multiply(2)).End(print())
	cancel[0] = p.Cancel

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	t := time.NewTimer(time.Second / 3)
	defer t.Stop()

ForLoop:
	for {
		select {
		case sig := <-signalChan:
			if sig == os.Interrupt {
				p.Cancel()
			}
		case <-p.Done():
			break ForLoop
		case <-t.C:
			p.Cancel()
		}
	}
}
